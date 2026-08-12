package policy

import (
	"regexp"
	"strings"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// StorageFiltersForList builds storage filters that enforce list resource patterns
// at query time (AND-ed with client filters). unrestricted means no extra filters.
// Patterns prefixed with "!" are Deny exclusions; others are Allow includes.
// Final query: (allow1 OR allow2 OR ...) AND (not deny1) AND (not deny2) ...
func (a *API) StorageFiltersForList(subject Subject, kind string) (unrestricted bool, filters []storageTY.Filter, err error) {
	unrestricted, patterns, err := a.ResourceNamesForList(subject, kind)
	if err != nil {
		return false, nil, err
	}
	if unrestricted {
		return true, nil, nil
	}
	if len(patterns) == 0 {
		return false, matchNothingFilters(), nil
	}

	allowPart, denyPart := splitAllowDenyPatterns(patterns)

	var out []storageTY.Filter

	// Positive allow scope (OR of pattern groups)
	if len(allowPart) > 0 {
		groups := make([][]storageTY.Filter, 0, len(allowPart))
		for _, p := range allowPart {
			name := p
			if hasResourceKind(p) {
				_, name = splitResource(p)
			}
			g := patternToFilterGroup(kind, name)
			if len(g) > 0 {
				groups = append(groups, g)
			}
		}
		if len(groups) == 0 {
			return false, matchNothingFilters(), nil
		}
		if len(groups) == 1 {
			out = append(out, groups[0]...)
		} else {
			out = append(out, storageTY.Filter{
				Operator: storageTY.OperatorOr,
				Value:    groups,
			})
		}
	} else if len(denyPart) == 0 {
		return false, matchNothingFilters(), nil
	}

	// Deny exclusions (AND)
	if len(denyPart) > 0 {
		out = append(out, denyOnlyFilters(kind, denyPart)...)
	}

	if len(out) == 0 {
		return false, matchNothingFilters(), nil
	}
	return false, out, nil
}

func matchNothingFilters() []storageTY.Filter {
	return []storageTY.Filter{{
		Key:      types.KeyID,
		Operator: storageTY.OperatorIn,
		Value:    []string{},
	}}
}

// denyOnlyFilters builds exclusion filters for allow-all-except-deny patterns.
// Each "!" name becomes a top-level AND filter meaning NOT that resource.
// Hierarchical paths use OperatorNor on the positive path match
// (NOT (GatewayID=gw AND NodeID=n …)).
func denyOnlyFilters(kind string, patterns []string) []storageTY.Filter {
	var filters []storageTY.Filter
	exactIDs := make([]string, 0)

	for _, p := range patterns {
		name := strings.TrimPrefix(p, "!")
		if name == "" {
			continue
		}
		if isIDKeyedKind(kind) && !strings.Contains(name, "*") {
			exactIDs = append(exactIDs, name)
			continue
		}
		excl := hierarchicalExcludeFilter(kind, name)
		if excl != nil {
			filters = append(filters, *excl)
		}
	}

	if len(exactIDs) > 0 {
		filters = append(filters, storageTY.Filter{
			Key:      types.KeyID,
			Operator: storageTY.OperatorNotIn,
			Value:    exactIDs,
		})
	}
	if len(filters) == 0 {
		// Could not encode denies; fail closed for list scope
		return matchNothingFilters()
	}
	return filters
}

// hierarchicalExcludeFilter returns a filter matching entities NOT under namePattern.
// Example: kind=field, name=mysensor.1 → NOR[GatewayID=mysensor, NodeID=1]
// Example: kind=node, name=mysensor.* → NOR[GatewayID=mysensor]
//
// Uses OperatorNor (NOT of the positive path) so it AND-s cleanly with Allow
// filters and does not rely on De Morgan OR of NotEqual (which collides with
// other $or groups in MongoDB).
func hierarchicalExcludeFilter(kind, namePattern string) *storageTY.Filter {
	switch kind {
	case policyTY.ResourceNode, policyTY.ResourceSource, policyTY.ResourceField:
		positive := patternToFilterGroup(kind, namePattern)
		if len(positive) == 0 {
			return nil
		}
		return &storageTY.Filter{
			Operator: storageTY.OperatorNor,
			Value:    positive,
		}
	default:
		if !strings.Contains(namePattern, "*") {
			return &storageTY.Filter{
				Key:      types.KeyID,
				Operator: storageTY.OperatorNotEqual,
				Value:    namePattern,
			}
		}
		// wildcards on id-keyed kinds: regex positive → not easily NOR-able; fail open nil
		// (caller fails closed if no filters)
		return nil
	}
}

func isIDKeyedKind(kind string) bool {
	switch kind {
	case policyTY.ResourceGateway, policyTY.ResourceTask, policyTY.ResourceSchedule,
		policyTY.ResourceHandler, policyTY.ResourceDashboard, policyTY.ResourceFirmware,
		policyTY.ResourceForwardPayload, policyTY.ResourceDataRepository,
		policyTY.ResourceVirtualDevice, policyTY.ResourceVirtualAssistant,
		policyTY.ResourceServiceToken, policyTY.ResourceUser, policyTY.ResourcePolicy:
		return true
	default:
		return false
	}
}

// patternToFilterGroup converts a resource name pattern into AND filters on entity fields.
// Supports trailing ".*" / "*" wildcards on hierarchical names.
func patternToFilterGroup(kind, namePattern string) []storageTY.Filter {
	namePattern = strings.TrimSpace(namePattern)
	if namePattern == "" || namePattern == "*" {
		return nil
	}

	// ID-keyed resources (gateway, task, schedule, ...)
	switch kind {
	case policyTY.ResourceGateway,
		policyTY.ResourceTask,
		policyTY.ResourceSchedule,
		policyTY.ResourceHandler,
		policyTY.ResourceDashboard,
		policyTY.ResourceFirmware,
		policyTY.ResourceForwardPayload,
		policyTY.ResourceDataRepository,
		policyTY.ResourceVirtualDevice,
		policyTY.ResourceVirtualAssistant,
		policyTY.ResourceServiceToken,
		policyTY.ResourceUser,
		policyTY.ResourcePolicy:
		return idPatternFilters(types.KeyID, namePattern)
	}

	// hierarchical device-tree resources
	parts := splitNamePattern(namePattern)
	switch kind {
	case policyTY.ResourceNode:
		return hierarchicalFilters(parts, []string{types.KeyGatewayID, types.KeyNodeID})
	case policyTY.ResourceSource:
		return hierarchicalFilters(parts, []string{types.KeyGatewayID, types.KeyNodeID, types.KeySourceID})
	case policyTY.ResourceField:
		return hierarchicalFilters(parts, []string{types.KeyGatewayID, types.KeyNodeID, types.KeySourceID, types.KeyFieldID})
	default:
		return idPatternFilters(types.KeyID, namePattern)
	}
}

func idPatternFilters(key, namePattern string) []storageTY.Filter {
	if strings.HasSuffix(namePattern, ".*") {
		prefix := strings.TrimSuffix(namePattern, ".*")
		return []storageTY.Filter{{
			Key:      key,
			Operator: storageTY.OperatorRegexCaseSensitive,
			Value:    "^" + regexp.QuoteMeta(prefix) + `($|\.)`,
		}}
	}
	if strings.HasSuffix(namePattern, "*") && !strings.HasSuffix(namePattern, ".*") {
		prefix := strings.TrimSuffix(namePattern, "*")
		return []storageTY.Filter{{
			Key:      key,
			Operator: storageTY.OperatorRegexCaseSensitive,
			Value:    "^" + regexp.QuoteMeta(prefix),
		}}
	}
	return []storageTY.Filter{{
		Key:      key,
		Operator: storageTY.OperatorEqual,
		Value:    namePattern,
	}}
}

// hierarchicalFilters maps name segments onto entity keys.
// A segment "*" or trailing ".*" stops further constraints (prefix scope).
func hierarchicalFilters(parts []string, keys []string) []storageTY.Filter {
	filters := make([]storageTY.Filter, 0, len(keys))
	for i, key := range keys {
		if i >= len(parts) {
			break
		}
		seg := parts[i]
		if seg == "*" {
			break
		}
		// last segment may be "foo*" style
		if strings.HasSuffix(seg, "*") {
			prefix := strings.TrimSuffix(seg, "*")
			filters = append(filters, storageTY.Filter{
				Key:      key,
				Operator: storageTY.OperatorRegexCaseSensitive,
				Value:    "^" + regexp.QuoteMeta(prefix),
			})
			break
		}
		filters = append(filters, storageTY.Filter{
			Key:      key,
			Operator: storageTY.OperatorEqual,
			Value:    seg,
		})
	}
	return filters
}

func splitNamePattern(name string) []string {
	// "home-gw.living-room.*" → ["home-gw", "living-room", "*"]
	name = strings.TrimSpace(name)
	if strings.HasSuffix(name, ".*") {
		base := strings.TrimSuffix(name, ".*")
		if base == "" {
			return []string{"*"}
		}
		parts := strings.Split(base, ".")
		return append(parts, "*")
	}
	return strings.Split(name, ".")
}
