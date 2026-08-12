package policy

import (
	"strings"
)

// MatchAction returns true if pattern matches action.
// "*" matches any action; otherwise an exact match, ignoring case and surrounding
// space so a hand written or UI supplied "Get" behaves like "get". Actions are a
// closed vocabulary, so folding case cannot widen access to a different verb.
//
// Resource *names* stay case sensitive: they are entity ids, compared exactly by
// storage, and folding them could merge two distinct entities.
func MatchAction(pattern, action string) bool {
	pattern = strings.ToLower(strings.TrimSpace(pattern))
	action = strings.ToLower(strings.TrimSpace(action))
	if pattern == "" || action == "" {
		return false
	}
	if pattern == "*" {
		return true
	}
	return pattern == action
}

// MatchResource returns true if pattern matches resource (same kind only).
// Resource format: "kind", "kind:name", or "*".
// For device-tree parent→child cascade (Allow or Deny), use the Cascade helpers.
func MatchResource(pattern, resource string) bool {
	if pattern == "" || resource == "" {
		return false
	}
	if pattern == "*" {
		return true
	}

	pKind, pName := splitResource(pattern)
	rKind, rName := splitResource(resource)

	if pKind != "*" && pKind != rKind {
		return false
	}

	// pattern is kind-only or kind:*
	if pName == "" || pName == "*" {
		return true
	}

	// resource is kind-only (list/create collection check): any pattern for this kind matches
	if rName == "" {
		return true
	}

	return nameMatchesPattern(pName, rName)
}

// MatchResourceAllowCascade matches an Allow pattern against a resource.
// Device-tree parent allows grant children under the same path:
//
//	Allow node:gw.n     → field:gw.n.s.f, source:gw.n.s, metric:gw.n…
//	Allow gateway:gw    → node/source/field/metric under gw
//	Allow node:gw.*     → all sources/fields under that gateway
//
// Named parent Allow also permits the child list API (filters scope rows).
// Kind-wide Allow on an ancestor (node / node:*) grants all of the child kind.
func MatchResourceAllowCascade(pattern, resource string) bool {
	if MatchResource(pattern, resource) {
		return true
	}
	return matchDeviceTreeCascade(pattern, resource, true)
}

// MatchResourceDenyCascade matches a Deny pattern against a resource.
// Device-tree parent denies block children under the same path (same hierarchy as Allow).
// Named Deny does not block the kind-only collection resource (list entry check);
// list rows are filtered separately.
func MatchResourceDenyCascade(pattern, resource string) bool {
	if pattern == "" || resource == "" {
		return false
	}
	if pattern == "*" {
		return true
	}

	pKind, pName := splitResource(pattern)
	rKind, rName := splitResource(resource)
	if pKind == "" || rKind == "" {
		return false
	}

	// Same kind (or pattern kind *)
	if pKind == "*" || pKind == rKind {
		if pName == "" || pName == "*" || pKind == "*" {
			return true
		}
		if rName == "" {
			return false // named deny does not block list API
		}
		return nameCoveredByPattern(pName, rName)
	}

	return matchDeviceTreeCascade(pattern, resource, false)
}

// matchDeviceTreeCascade: parent kind pattern covers child kind resource by path.
// allowCollection: when true, named parent patterns match kind-only child resources (list).
func matchDeviceTreeCascade(pattern, resource string, allowCollection bool) bool {
	pKind, pName := splitResource(pattern)
	rKind, rName := splitResource(resource)
	if !deviceTreeCascadesTo(pKind, rKind) {
		return false
	}
	if pName == "" || pName == "*" {
		// kind-wide parent → all children of that kind
		return true
	}
	if rName == "" {
		return allowCollection
	}
	return nameCoveredByPattern(pName, rName)
}

// deviceTreeCascadesTo reports whether patternKind is a device-tree ancestor of resourceKind.
// Same kind is handled by the caller; this is parent→child only.
// Hierarchy: gateway → node → source → field/metric.
func deviceTreeCascadesTo(patternKind, resourceKind string) bool {
	if patternKind == "" || resourceKind == "" || patternKind == resourceKind {
		return false
	}
	ancestors := map[string][]string{
		"node":   {"gateway"},
		"source": {"gateway", "node"},
		"field":  {"gateway", "node", "source"},
		"metric": {"gateway", "node", "source", "field"},
	}
	for _, a := range ancestors[resourceKind] {
		if a == patternKind {
			return true
		}
	}
	return false
}

// nameCoveredByPattern: pattern name matches resource name, or is a strict
// hierarchical parent (mysensor.1 covers mysensor.1.s.f).
// Segment-safe: mysensor.1 does not cover mysensor.10.
func nameCoveredByPattern(patternName, resourceName string) bool {
	if nameMatchesPattern(patternName, resourceName) {
		return true
	}
	// exact parent path (no wildcards): "a.b" covers "a.b.c.d"
	if strings.Contains(patternName, "*") {
		return false
	}
	if resourceName == patternName {
		return true
	}
	return strings.HasPrefix(resourceName, patternName+".")
}

func nameMatchesPattern(pName, rName string) bool {
	if pName == rName {
		return true
	}
	// trailing wildcard: "home-gw.living-room.*"
	if strings.HasSuffix(pName, ".*") {
		prefix := strings.TrimSuffix(pName, ".*")
		if rName == prefix {
			return true
		}
		return strings.HasPrefix(rName, prefix+".")
	}
	if strings.HasSuffix(pName, "*") {
		prefix := strings.TrimSuffix(pName, "*")
		return strings.HasPrefix(rName, prefix)
	}
	return false
}

// FormatResource builds "kind" or "kind:name".
func FormatResource(kind, name string) string {
	if kind == "" {
		return ""
	}
	if name == "" || name == "*" {
		return kind
	}
	return kind + ":" + name
}

func splitResource(resource string) (kind, name string) {
	parts := strings.SplitN(resource, ":", 2)
	kind = strings.ToLower(strings.TrimSpace(parts[0]))
	if len(parts) == 2 {
		name = strings.TrimSpace(parts[1])
	}
	return kind, name
}
