package policy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
)

// maxBodyPeekBytes caps how much of a write body is buffered for authorization.
const maxBodyPeekBytes = 4 << 20 // 4 MiB

// bodyTarget carries every identity field a write payload may use to name its
// target. Entity payloads across the api are uniform: "id" for id-keyed kinds and
// the gateway/node/source/field chain for the device tree.
type bodyTarget struct {
	ID        string `json:"id"`
	GatewayID string `json:"gatewayId"`
	NodeID    string `json:"nodeId"`
	SourceID  string `json:"sourceId"`
	FieldID   string `json:"fieldId"`
}

// name returns the policy resource name this payload targets, for the given kind.
func (b bodyTarget) name(kind string) string {
	switch kind {
	case policyTY.ResourceNode:
		if path := joinIDs(b.GatewayID, b.NodeID); path != "" {
			return path
		}
	case policyTY.ResourceSource:
		if path := joinIDs(b.GatewayID, b.NodeID, b.SourceID); path != "" {
			return path
		}
	case policyTY.ResourceField:
		if path := joinIDs(b.GatewayID, b.NodeID, b.SourceID, b.FieldID); path != "" {
			return path
		}
	}
	return b.ID
}

// AuthorizeBodyTargets authorizes the objects named in a write payload.
//
// The path-based check in the middleware can only see the kind for collection
// endpoints - POST /api/gateway, POST /api/gateway/enable and DELETE /api/gateway
// all carry their targets in the body. A kind-only check is deliberately permissive
// (it answers "may you reach this endpoint?"), so without this second pass a grant
// on one named object would authorize writes to every object of that kind.
//
// Two payload shapes cover the whole api:
//
//	{"id": "gw1", ...}        one entity  -> check kind:gw1
//	["gw1", "gw2"]            bulk ids    -> check kind:gw1 and kind:gw2
//
// A payload that names no target (create with a server generated id) requires a
// kind-wide grant instead. Anything else falls back to the kind-level decision
// already made by the caller.
//
// The body is restored so the handler can read it.
func (a *API) AuthorizeBodyTargets(subject Subject, r *http.Request, access *RequestAccess) error {
	if access == nil || access.Kind == "" || access.Skip {
		return nil
	}
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
	default:
		return nil
	}
	// the path already named the target and it has been checked. This is also the
	// only place uploads arrive (POST /api/firmware/upload/{id}), so no large or
	// streaming body is ever buffered here.
	if access.Name != "" {
		return nil
	}

	// Deliberately not keyed on Content-Type: handlers json.Unmarshal the body
	// whatever the header says, so trusting it would let a client skip this check
	// by sending "text/plain".
	body, err := a.peekBody(r)
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return nil
	}

	switch firstJSONToken(body) {
	case '[':
		return a.authorizeIDList(subject, body, access)
	case '{':
		return a.authorizeEntity(subject, body, access)
	default:
		// not a json document: the handler cannot decode it either, so it cannot
		// write anything. The kind level decision stands.
		return nil
	}
}

// authorizeIDList handles bulk id payloads (enable / disable / reload / delete).
func (a *API) authorizeIDList(subject Subject, body []byte, access *RequestAccess) error {
	var ids []string
	if err := json.Unmarshal(body, &ids); err != nil {
		// not a list of ids (e.g. a list of objects): the kind level decision stands.
		// Endpoints that take structured lists of targets are authorized explicitly
		// (see AuthorizeActionRequest).
		return nil
	}
	if len(ids) == 0 {
		return nil
	}
	// for id-keyed kinds the id *is* the policy name, so no lookup is needed
	resolve := !isIDKeyedKind(access.Kind)
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if err := a.allowedTarget(subject, access.Action, access.Kind, id, resolve); err != nil {
			return err
		}
	}
	return nil
}

// authorizeEntity handles single entity payloads.
func (a *API) authorizeEntity(subject Subject, body []byte, access *RequestAccess) error {
	target := bodyTarget{}
	if err := json.Unmarshal(body, &target); err != nil {
		// malformed or unexpected shape: the handler will reject it. Require the
		// stronger kind-wide grant rather than trusting the kind level decision.
		return a.AllowedKindWide(subject, access.Action, access.Kind)
	}

	if access.Kind == policyTY.ResourceUser {
		if err := a.authorizeUserPrivilegedFields(subject, body, access); err != nil {
			return err
		}
	}

	name := strings.TrimSpace(target.name(access.Kind))
	if name == "" {
		// no target to check (create with a server generated id): a grant on some
		// other named object must not be enough
		return a.AllowedKindWide(subject, access.Action, access.Kind)
	}
	// A client supplied name is not necessarily an existing object, so only resolve
	// storage ids for the device tree, where the payload may carry a bare uuid.
	return a.allowedTarget(subject, access.Action, access.Kind, name, !isIDKeyedKind(access.Kind))
}

// authorizeUserPrivilegedFields requires a kind-wide user grant to change
// policies or disabled. Assigning the built-in admin policy additionally
// requires that the caller already holds equivalent full access.
func (a *API) authorizeUserPrivilegedFields(subject Subject, body []byte, access *RequestAccess) error {
	raw := map[string]json.RawMessage{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return a.AllowedKindWide(subject, access.Action, policyTY.ResourceUser)
	}
	if _, ok := raw["policies"]; ok {
		if err := a.AllowedKindWide(subject, access.Action, policyTY.ResourceUser); err != nil {
			return fmt.Errorf("update user policies: %w", err)
		}
	}
	if _, ok := raw["disabled"]; ok {
		if err := a.AllowedKindWide(subject, access.Action, policyTY.ResourceUser); err != nil {
			return fmt.Errorf("update user disabled: %w", err)
		}
	}
	if userPayloadAssignsAdmin(raw) && !a.subjectHoldsAdmin(subject) {
		return fmt.Errorf("assigning admin policy: %w", ErrAccessDenied)
	}
	return nil
}

func userPayloadAssignsAdmin(raw map[string]json.RawMessage) bool {
	msg, ok := raw["policies"]
	if !ok {
		return false
	}
	var policies []string
	if err := json.Unmarshal(msg, &policies); err != nil {
		return true
	}
	for _, id := range policies {
		if strings.TrimSpace(id) == policyTY.PolicyAdmin {
			return true
		}
	}
	return false
}

func (a *API) subjectHoldsAdmin(subject Subject) bool {
	user, err := a.activeUser(subject)
	if err != nil {
		return false
	}
	return a.policiesAllow(user.Policies, policyTY.ActionAll, policyTY.ResourceAll)
}

// allowedTarget checks action on kind:<name>. When resolve is set and the name is a
// storage id, it is translated to the business name used in policies
// (uuid -> gatewayId.nodeId...).
func (a *API) allowedTarget(subject Subject, action, kind, name string, resolve bool) error {
	if resolve && !strings.Contains(name, ".") {
		if biz, err := a.ResolveBusinessName(kind, name); err == nil && biz != "" {
			name = biz
		}
	}
	resource := FormatResource(kind, name)
	if err := a.Allowed(subject, action, resource); err != nil {
		return fmt.Errorf("%s on %s: %w", action, resource, err)
	}
	return nil
}

// peekBody buffers the request body for inspection and puts it back for the handler.
func (a *API) peekBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyPeekBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxBodyPeekBytes {
		return nil, fmt.Errorf("request payload too large")
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

// firstJSONToken returns the first meaningful byte of a json document.
func firstJSONToken(body []byte) byte {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return 0
	}
	return trimmed[0]
}
