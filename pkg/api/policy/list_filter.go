package policy

import (
	"reflect"
	"strings"

	dashboardTY "github.com/mycontroller-org/server/v2/pkg/types/dashboard"
	dataRepoTY "github.com/mycontroller-org/server/v2/pkg/types/data_repository"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	firmwareTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	fwdPayloadTY "github.com/mycontroller-org/server/v2/pkg/types/forward_payload"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	sourceTY "github.com/mycontroller-org/server/v2/pkg/types/source"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	vdTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_device"
	gatewayTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	vaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/types"
)

// ItemAllowedForList reports whether one item's business name matches list patterns.
// Patterns may include "!" deny exclusions from ResourceNamesForList.
func ItemAllowedForList(kind string, patterns []string, item interface{}) bool {
	name := BusinessName(kind, item)
	if name == "" {
		return false
	}
	resource := FormatResource(kind, name)
	allowPart, denyPart := splitAllowDenyPatterns(patterns)

	for _, p := range denyPart {
		namePat := strings.TrimPrefix(p, "!")
		if namePat == "" {
			continue
		}
		if hasResourceKind(namePat) {
			_, namePat = splitResource(namePat)
		}
		if nameCoveredByPattern(namePat, name) {
			return false
		}
	}

	if len(allowPart) == 0 {
		// deny-only: allow all except denied
		return len(denyPart) > 0
	}

	for _, p := range allowPart {
		// patterns are name-only from ResourceNamesForList; also accept full resource strings
		pattern := p
		if _, n := splitResource(p); n == "" && p != "*" {
			// kind-only stored as name by mistake
			pattern = FormatResource(kind, p)
		} else if !hasResourceKind(p) {
			pattern = FormatResource(kind, p)
		}
		if MatchResource(pattern, resource) {
			return true
		}
	}
	return false
}

func hasResourceKind(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			return true
		}
	}
	return false
}

// BusinessName returns the policy resource name for an entity (not storage UUID when a path exists).
func BusinessName(kind string, item interface{}) string {
	if item == nil {
		return ""
	}
	// pointer to struct
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return ""
		}
		item = v.Elem().Interface()
	}

	switch kind {
	case policyTY.ResourceGateway:
		if e, ok := item.(gatewayTY.Config); ok {
			return e.ID
		}
	case policyTY.ResourceNode:
		if e, ok := item.(nodeTY.Node); ok {
			return joinIDs(e.GatewayID, e.NodeID)
		}
	case policyTY.ResourceSource:
		if e, ok := item.(sourceTY.Source); ok {
			return joinIDs(e.GatewayID, e.NodeID, e.SourceID)
		}
	case policyTY.ResourceField:
		if e, ok := item.(fieldTY.Field); ok {
			return joinIDs(e.GatewayID, e.NodeID, e.SourceID, e.FieldID)
		}
	case policyTY.ResourceTask:
		if e, ok := item.(taskTY.Config); ok {
			return e.ID
		}
	case policyTY.ResourceSchedule:
		if e, ok := item.(schedulerTY.Config); ok {
			return e.ID
		}
	case policyTY.ResourceHandler:
		if e, ok := item.(handlerTY.Config); ok {
			return e.ID
		}
	case policyTY.ResourceDashboard:
		if e, ok := item.(dashboardTY.Config); ok {
			return e.ID
		}
	case policyTY.ResourceFirmware:
		if e, ok := item.(firmwareTY.Firmware); ok {
			return e.ID
		}
	case policyTY.ResourceForwardPayload:
		if e, ok := item.(fwdPayloadTY.Config); ok {
			return e.ID
		}
	case policyTY.ResourceDataRepository:
		if e, ok := item.(dataRepoTY.Config); ok {
			return e.ID
		}
	case policyTY.ResourceVirtualDevice:
		if e, ok := item.(vdTY.VirtualDevice); ok {
			return e.ID
		}
	case policyTY.ResourceVirtualAssistant:
		if e, ok := item.(vaTY.Config); ok {
			return e.ID
		}
	}

	// generic: ID field via reflection
	rv := reflect.ValueOf(item)
	if rv.Kind() == reflect.Struct {
		f := rv.FieldByName("ID")
		if f.IsValid() && f.Kind() == reflect.String {
			return f.String()
		}
	}
	return ""
}

func joinIDs(parts ...string) string {
	out := ""
	for i, p := range parts {
		if p == "" {
			continue
		}
		if i > 0 && out != "" {
			out += "."
		}
		out += p
	}
	return out
}
