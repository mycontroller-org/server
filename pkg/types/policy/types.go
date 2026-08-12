package policy

import (
	"strings"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

// Built-in policy IDs
const (
	PolicyAdmin     = "admin"
	PolicyReadWrite = "readwrite"
	PolicyReadOnly  = "readonly"
)

// Effects
const (
	EffectAllow = "Allow"
	EffectDeny  = "Deny" // explicit deny wins over allow
)

// Common actions
const (
	ActionGet     = "get"
	ActionList    = "list"
	ActionCreate  = "create"
	ActionUpdate  = "update"
	ActionDelete  = "delete"
	ActionEnable  = "enable"
	ActionDisable = "disable"
	ActionReload  = "reload"
	ActionAction  = "action" // execute resource action
	ActionAll     = "*"
)

// Common resource kinds (used in resource strings as kind:name)
const (
	ResourceGateway          = "gateway"
	ResourceNode             = "node"
	ResourceSource           = "source"
	ResourceField            = "field"
	ResourceTask             = "task"
	ResourceSchedule         = "schedule"
	ResourceHandler          = "handler"
	ResourceDashboard        = "dashboard"
	ResourceFirmware         = "firmware"
	ResourceForwardPayload   = "forwardpayload"
	ResourceDataRepository   = "datarepository"
	ResourceVirtualDevice    = "virtualdevice"
	ResourceVirtualAssistant = "virtualassistant"
	ResourceServiceToken     = "servicetoken"
	ResourceSettings         = "settings"
	ResourceBackup           = "backup"
	ResourceUser             = "user"
	ResourcePolicy           = "policy"
	ResourceMetric           = "metric"
	ResourceAction           = "action"
	ResourceStatus           = "status"
	ResourceQuickID          = "quickid"
	ResourceAll              = "*"
)

// FormatSettingsResource builds the resource string for one settings document,
// e.g. "settings:system_settings".
func FormatSettingsResource(key string) string {
	return ResourceSettings + ":" + key
}

// NormalizeKind maps an api path segment or a storage entity name to a policy kind.
// Single source of truth: api paths use "forwardpayload", storage uses
// "forward_payload", policies use ResourceForwardPayload. Keeping one mapping
// prevents a kind from silently losing its access-control scope.
func NormalizeKind(value string) string {
	s := strings.ToLower(strings.TrimSpace(value))
	switch s {
	case "forward_payload", "forwardpayload":
		return ResourceForwardPayload
	case "data_repository", "datarepository":
		return ResourceDataRepository
	case "virtual_device", "virtualdevice":
		return ResourceVirtualDevice
	case "virtual_assistant", "virtualassistant":
		return ResourceVirtualAssistant
	case "service_token", "servicetoken":
		return ResourceServiceToken
	default:
		return s
	}
}

// Policy is a named permission document attached to users (and optionally used when resolving access).
type Policy struct {
	ID          string               `json:"id" yaml:"id"`
	Description string               `json:"description" yaml:"description"`
	System      bool                 `json:"system" yaml:"system"` // built-in; protect from delete
	Statements  []Statement          `json:"statements" yaml:"statements"`
	Labels      cmap.CustomStringMap `json:"labels" yaml:"labels"`
	ModifiedOn  time.Time            `json:"modifiedOn" yaml:"modifiedOn"`
}

// Statement grants actions on resources.
// Resource format: "kind" | "kind:name" | "kind:name.*" | "*"
// Examples: "field:home-gw.living-room.dht.temperature", "node:home-gw.*", "gateway:*", "*"
type Statement struct {
	Effect    string   `json:"effect" yaml:"effect"` // Allow or Deny (Deny wins)
	Actions   []string `json:"actions" yaml:"actions"`
	Resources []string `json:"resources" yaml:"resources"`
}
