package policy

import (
	"errors"
	"strings"
	"time"

	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	svcTokenTY "github.com/mycontroller-org/server/v2/pkg/types/service_token"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
)

var (
	errCacheNotReady = errors.New("access control cache not ready")
	ErrUserDisabled  = errors.New("user is disabled")
	ErrUserNotFound  = errors.New("user not found")
	ErrTokenExpired  = errors.New("service token expired")
	ErrTokenNotFound = errors.New("service token not found")
	ErrAccessDenied  = errors.New("access denied")
)

// Subject is the authenticated principal for an access check.
type Subject struct {
	UserID         string
	ServiceTokenID string // raw Token.ID from JWT; empty for interactive login
}

// Allowed reports whether the subject may perform action on resource.
// resource should be FormatResource(kind, name) e.g. "field:gw.n.s.f" or "gateway" for list-all.
//
// Rules:
//  1. User must exist and not be disabled
//  2. If service token: must exist, not expired, belong to user
//  3. User policies must allow (ceiling)
//  4. If token has restrictions, they must also allow (can only lower)
func (a *API) Allowed(subject Subject, action, resource string) error {
	user, err := a.activeUser(subject)
	if err != nil {
		return err
	}
	token, err := a.activeToken(subject)
	if err != nil {
		return err
	}

	if !a.policiesAllow(user.Policies, action, resource) {
		return ErrAccessDenied
	}

	// Token restrictions (optional lower bound): can only narrow further
	if token != nil && (len(token.Actions) > 0 || len(token.Resources) > 0) {
		if !restrictionsAllow(token.Actions, token.Resources, action, resource) {
			return ErrAccessDenied
		}
	}

	return nil
}

// activeUser loads the subject's user and verifies it can be used for authorization.
// A user with no attached policies has no access at all.
func (a *API) activeUser(subject Subject) (*userTY.User, error) {
	if subject.UserID == "" {
		return nil, ErrUserNotFound
	}
	user, err := a.cache.GetUser(subject.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	if user.Disabled {
		return nil, ErrUserDisabled
	}
	if len(user.Policies) == 0 {
		return nil, ErrAccessDenied
	}
	return user, nil
}

// activeToken loads the subject's service token, if the request presented one.
// Returns (nil, nil) for an interactive login.
func (a *API) activeToken(subject Subject) (*svcTokenTY.ServiceToken, error) {
	if subject.ServiceTokenID == "" {
		return nil, nil
	}
	token, err := a.cache.GetToken(subject.ServiceTokenID)
	if err != nil {
		return nil, ErrTokenNotFound
	}
	if token.UserID != subject.UserID {
		return nil, ErrAccessDenied
	}
	if err := validateTokenExpiry(token); err != nil {
		return nil, err
	}
	return token, nil
}

// AllowedKindWide reports whether the subject may perform action on *any* object of
// kind, i.e. it holds a kind-wide grant ("*", "kind" or "kind:*"), not merely a
// grant on some named object.
//
// Needed for writes whose target cannot be named in advance - creating an object
// with a server generated id. Without this, "Allow update on user:self" would be
// enough to create new users, because a kind-only resource check is deliberately
// permissive (it answers "may you reach this collection endpoint?").
func (a *API) AllowedKindWide(subject Subject, action, kind string) error {
	user, err := a.activeUser(subject)
	if err != nil {
		return err
	}
	token, err := a.activeToken(subject)
	if err != nil {
		return err
	}

	if !a.policiesAllowKind(user.Policies, action, kind, true) {
		return ErrAccessDenied
	}
	// A token restriction naming individual objects cannot satisfy a kind-wide check
	if token != nil && (len(token.Actions) > 0 || len(token.Resources) > 0) {
		if len(token.Actions) > 0 && !anyActionMatch(token.Actions, action) {
			return ErrAccessDenied
		}
		if len(token.Resources) > 0 && !resourcesCoverKindWide(token.Resources, kind) {
			return ErrAccessDenied
		}
	}
	return nil
}

// resourcesCoverKindWide reports whether any resource pattern covers the whole kind.
func resourcesCoverKindWide(resources []string, kind string) bool {
	for _, res := range resources {
		if res == "*" {
			return true
		}
		k, name := splitResource(res)
		if k != kind && k != "*" {
			continue
		}
		if name == "" || name == "*" {
			return true
		}
	}
	return false
}

// EnsureUserActive loads user from cache and verifies not disabled (for auth middleware).
func (a *API) EnsureUserActive(userID string) (*userTY.User, error) {
	user, err := a.cache.GetUser(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	if user.Disabled {
		return nil, ErrUserDisabled
	}
	return user, nil
}

// EnsureServiceTokenActive validates token still valid for requests.
func (a *API) EnsureServiceTokenActive(userID, tokenID string) error {
	if tokenID == "" {
		return nil
	}
	token, err := a.cache.GetToken(tokenID)
	if err != nil {
		return ErrTokenNotFound
	}
	if token.UserID != userID {
		return ErrAccessDenied
	}
	return validateTokenExpiry(token)
}

func validateTokenExpiry(token *svcTokenTY.ServiceToken) error {
	if token.NeverExpire {
		return nil
	}
	if token.ExpiresOn.IsZero() {
		return ErrTokenExpired
	}
	// ExpiresOn is date-only; treat as end of that calendar day in local time is complex -
	// CustomDate Before uses time.Time; compare with start of tomorrow conceptually.
	// Keep same semantics as login: ExpiresOn.Before(time.Now()) means expired.
	if token.ExpiresOn.Before(time.Now()) {
		return ErrTokenExpired
	}
	return nil
}

// policiesAllow evaluates all attached policies.
// Explicit Deny matching action+resource always wins over Allow (IAM-style).
//
// For kind-only resources (list entry checks like "field"), a Deny on a *named*
// resource (e.g. field:site-a.secret) does not block the whole list call;
// only Deny on kind-wide patterns (field, field:*, *) blocks list for that kind.
// Named Deny is applied when checking a specific name (get by id / list filters).
func (a *API) policiesAllow(policyIDs []string, action, resource string) bool {
	kind, name := splitResource(resource)
	if name == "" && kind != "" && kind != "*" {
		return a.policiesAllowKind(policyIDs, action, kind, false)
	}
	return evaluateStatements(a.collectStatements(policyIDs), action, resource)
}

func (a *API) collectStatements(policyIDs []string) []policyTY.Statement {
	out := make([]policyTY.Statement, 0)
	for _, id := range policyIDs {
		p, err := a.cache.GetPolicy(id)
		if err != nil {
			continue
		}
		out = append(out, p.Statements...)
	}
	return out
}

// policiesAllowKind evaluates a kind-only resource (a collection, not one object).
//
// requireKindWide=false (collection reachability, e.g. "may I call list / may I POST
// to this endpoint?"): a grant on a single named object is enough, because the
// object-level check happens afterwards on the concrete target.
//
// requireKindWide=true: only "*", "kind" or "kind:*" grants count. Used where no
// concrete target exists to check later (create with a server generated id).
//
// Device-tree: Allow node:gw.* also permits list of source/field (rows filtered by path).
// Kind-wide Deny on an ancestor (node / node:*) blocks list of descendants.
func (a *API) policiesAllowKind(policyIDs []string, action, kind string, requireKindWide bool) bool {
	var allowed, denyAll bool
	for _, st := range a.collectStatements(policyIDs) {
		if !anyActionMatch(st.Actions, action) {
			continue
		}
		effect := normalizeEffect(st.Effect)
		for _, res := range st.Resources {
			if res == "*" {
				if effect == policyTY.EffectDeny {
					denyAll = true
				} else {
					allowed = true
				}
				continue
			}
			k, n := splitResource(res)

			// Parent device-tree kinds: Allow cascade / kind-wide Deny cascade
			if k != kind && k != "*" && deviceTreeCascadesTo(k, kind) {
				if effect == policyTY.EffectDeny && (n == "" || n == "*") {
					denyAll = true
				} else if effect == policyTY.EffectAllow && (!requireKindWide || n == "" || n == "*") {
					// named or kind-wide parent Allow → may list child kind (filters apply)
					allowed = true
				}
				// named parent Deny does not block the list API itself
				continue
			}

			if k != kind && k != "*" {
				continue
			}
			// kind-wide pattern
			if n == "" || n == "*" || k == "*" {
				if effect == policyTY.EffectDeny {
					denyAll = true
				} else {
					allowed = true
				}
				continue
			}
			// named Allow still permits calling list (filters will scope rows),
			// but never satisfies a kind-wide requirement
			if effect == policyTY.EffectAllow && !requireKindWide {
				allowed = true
			}
			// named Deny does not block the list API itself
		}
	}
	if denyAll {
		return false
	}
	return allowed
}

func normalizeEffect(effect string) string {
	if effect == "" {
		return policyTY.EffectAllow
	}
	// accept common casings from UI / hand-edited YAML
	switch strings.ToLower(strings.TrimSpace(effect)) {
	case "deny":
		return policyTY.EffectDeny
	case "allow":
		return policyTY.EffectAllow
	default:
		return effect
	}
}

func statementMatches(st policyTY.Statement, action, resource string) bool {
	if !anyActionMatch(st.Actions, action) {
		return false
	}
	// Device-tree cascade for both effects (Deny still wins in evaluateStatements)
	if normalizeEffect(st.Effect) == policyTY.EffectDeny {
		return anyResourceMatchDenyCascade(st.Resources, resource)
	}
	return anyResourceMatchAllowCascade(st.Resources, resource)
}

func evaluateStatements(statements []policyTY.Statement, action, resource string) bool {
	var allowed, denied bool
	for _, st := range statements {
		if !statementMatches(st, action, resource) {
			continue
		}
		switch normalizeEffect(st.Effect) {
		case policyTY.EffectDeny:
			denied = true
		case policyTY.EffectAllow:
			allowed = true
		}
	}
	if denied {
		return false
	}
	return allowed
}

// statementsAllow is kept for tests: true if any Allow matches and no Deny matches.
func statementsAllow(statements []policyTY.Statement, action, resource string) bool {
	return evaluateStatements(statements, action, resource)
}

func restrictionsAllow(actions, resources []string, action, resource string) bool {
	// empty actions in restriction means all actions (still under user ceiling)
	if len(actions) > 0 && !anyActionMatch(actions, action) {
		return false
	}
	// token resource limits use the same device-tree Allow cascade as policies
	if len(resources) > 0 && !anyResourceMatchAllowCascade(resources, resource) {
		return false
	}
	return true
}

func anyActionMatch(patterns []string, action string) bool {
	for _, p := range patterns {
		if MatchAction(p, action) {
			return true
		}
	}
	return false
}

func anyResourceMatchAllowCascade(patterns []string, resource string) bool {
	for _, p := range patterns {
		if MatchResourceAllowCascade(p, resource) {
			return true
		}
	}
	return false
}

func anyResourceMatchDenyCascade(patterns []string, resource string) bool {
	for _, p := range patterns {
		if MatchResourceDenyCascade(p, resource) {
			return true
		}
	}
	return false
}

// ResourceNamesForList returns whether list is unrestricted for this kind, and optional name patterns
// from user policies (intersected with token resources). Used for filtered list queries.
// unrestricted=true means no name filter needed.
func (a *API) ResourceNamesForList(subject Subject, kind string) (unrestricted bool, patterns []string, err error) {
	if err := a.Allowed(subject, policyTY.ActionList, FormatResource(kind, "")); err != nil {
		// also try kind:*
		if err2 := a.Allowed(subject, policyTY.ActionList, FormatResource(kind, "*")); err2 != nil {
			return false, nil, err
		}
	}

	user, err := a.cache.GetUser(subject.UserID)
	if err != nil {
		return false, nil, ErrUserNotFound
	}

	// Collect allow/deny name patterns for this kind (list action)
	var allowPatterns, denyPatterns []string
	hasWildcard := false
	denyAll := false
	for _, pid := range user.Policies {
		p, err := a.cache.GetPolicy(pid)
		if err != nil {
			continue
		}
		for _, st := range p.Statements {
			if !anyActionMatch(st.Actions, policyTY.ActionList) {
				continue
			}
			effect := normalizeEffect(st.Effect)
			for _, res := range st.Resources {
				if res == "*" {
					if effect == policyTY.EffectDeny {
						denyAll = true
					} else {
						hasWildcard = true
					}
					continue
				}
				k, name := splitResource(res)

				// Ancestor kinds (gateway/node/source → field list, etc.)
				if k != kind && k != "*" && deviceTreeCascadesTo(k, kind) {
					if effect == policyTY.EffectDeny {
						if name == "" || name == "*" {
							denyAll = true
						} else {
							// path exclude for children under denied parent
							denyPatterns = append(denyPatterns, name)
						}
						continue
					}
					// Allow parent path → allow child rows under that path
					if name == "" || name == "*" {
						hasWildcard = true
					} else {
						allowPatterns = append(allowPatterns, name)
					}
					continue
				}

				if k != kind && k != "*" {
					continue
				}
				if name == "" || name == "*" {
					if effect == policyTY.EffectDeny {
						denyAll = true
					} else {
						hasWildcard = true
					}
					continue
				}
				if effect == policyTY.EffectDeny {
					denyPatterns = append(denyPatterns, name)
				} else {
					allowPatterns = append(allowPatterns, name)
				}
			}
		}
	}

	if denyAll {
		// kind-wide deny → no rows
		return false, nil, nil
	}

	// Drop allow patterns fully covered by an exact deny (optional tidy)
	if len(denyPatterns) > 0 && len(allowPatterns) > 0 {
		allowPatterns = subtractPatterns(allowPatterns, denyPatterns)
	}

	// Build combined pattern list:
	//   allow names as-is, deny names as "!name"
	// StorageFiltersForList applies allow (OR) AND deny exclusions.
	if hasWildcard {
		if len(denyPatterns) == 0 {
			unrestricted = true
			patterns = nil
		} else {
			// Allow all of kind except deny names
			unrestricted = false
			patterns = encodeDenyOnlyPatterns(denyPatterns)
		}
	} else {
		unrestricted = false
		patterns = append([]string{}, allowPatterns...)
		// Critical: Allow node:mysensor.* + Deny node:mysensor.1 must keep both
		if len(denyPatterns) > 0 {
			patterns = append(patterns, encodeDenyOnlyPatterns(denyPatterns)...)
		}
	}

	// Intersect with token restrictions (token can only narrow)
	if subject.ServiceTokenID != "" {
		token, err := a.cache.GetToken(subject.ServiceTokenID)
		if err != nil {
			return false, nil, ErrTokenNotFound
		}
		if len(token.Resources) > 0 {
			tokenPatterns, tokenWild := tokenNamePatternsForKind(token.Resources, kind)
			if !tokenWild {
				unrestricted = false
				allowPart, denyPart := splitAllowDenyPatterns(patterns)
				switch {
				case len(tokenPatterns) == 0:
					// token names resources, none of which reach this kind -> no rows
					allowPart = nil
				case hasWildcard && len(allowPart) == 0:
					// user policies allow the whole kind: token limits are the effective scope
					allowPart = tokenPatterns
				default:
					allowPart = intersectPatterns(allowPart, tokenPatterns)
				}
				if len(allowPart) == 0 {
					// nothing allowed: a deny-only pattern list would read as
					// "everything except ..." downstream, so drop it too
					denyPart = nil
				}
				patterns = append(allowPart, denyPart...)
			}
		}
	}

	return unrestricted, patterns, nil
}

// tokenNamePatternsForKind collects the name patterns a service token allows for kind.
// Mirrors the policy loop above, including the device-tree parent cascade
// (token resource "gateway:gw" reaches node/source/field/metric rows under gw),
// so list scoping matches what Allowed() permits for a single resource.
func tokenNamePatternsForKind(resources []string, kind string) (patterns []string, wildcard bool) {
	for _, res := range resources {
		if res == "*" {
			return nil, true
		}
		k, name := splitResource(res)
		if k != kind && k != "*" && !deviceTreeCascadesTo(k, kind) {
			continue
		}
		if name == "" || name == "*" {
			return nil, true
		}
		patterns = append(patterns, name)
	}
	return patterns, false
}

func splitAllowDenyPatterns(patterns []string) (allow, deny []string) {
	for _, p := range patterns {
		if len(p) > 0 && p[0] == '!' {
			deny = append(deny, p)
		} else {
			allow = append(allow, p)
		}
	}
	return allow, deny
}

// encodeDenyOnlyPatterns marks patterns as exclusions for StorageFiltersForList ("all except").
func encodeDenyOnlyPatterns(deny []string) []string {
	out := make([]string, 0, len(deny))
	for _, d := range deny {
		out = append(out, "!"+d)
	}
	return out
}

func subtractPatterns(allow, deny []string) []string {
	if len(deny) == 0 {
		return allow
	}
	out := make([]string, 0, len(allow))
	for _, a := range allow {
		denied := false
		for _, d := range deny {
			if a == d || MatchResource(FormatResource("x", d), FormatResource("x", a)) {
				denied = true
				break
			}
		}
		if !denied {
			out = append(out, a)
		}
	}
	return out
}

// intersectPatterns keeps only name patterns allowed by both sides
// (a = user policy scope, b = service token scope). When one pattern covers the
// other, the narrower one survives. No overlap means no access, so an empty
// result is a valid answer and must be treated as "no rows" by the caller.
func intersectPatterns(a, b []string) []string {
	if len(a) == 0 || len(b) == 0 {
		return nil
	}
	out := make([]string, 0)
	seen := make(map[string]struct{})
	keep := func(p string) {
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	for _, tokenPattern := range b {
		for _, userPattern := range a {
			switch {
			case tokenPattern == userPattern, nameCoveredByPattern(userPattern, tokenPattern):
				keep(tokenPattern) // token side is equal or narrower
			case nameCoveredByPattern(tokenPattern, userPattern):
				keep(userPattern) // user side is narrower
			}
		}
	}
	return out
}
