package types

import "strings"

// common resource actions
const (
	ActionEnable  = "enable"
	ActionDisable = "disable"
	ActionReload  = "reload"
	ActionDelete  = "delete"
	ActionToggle  = "toggle"

	KeyAction = "action"
)

// GetAction parse and rename if required
func GetAction(action string) string {
	action = strings.ToLower(action)
	switch action {
	case "true":
		return ActionEnable
	case "false":
		return ActionDisable
	default:
		return action
	}
}
