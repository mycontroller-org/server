package message

// Commands
const (
	TypeNone         = ""
	TypeSet          = "set"
	TypeRequest      = "request"
	TypePresentation = "presentation"
	TypeInternal     = "internal"
	TypeStream       = "stream"
)

// Sub command options
const (
	CommandNone     = "none"
	CommandReboot   = "reboot"
	CommandReset    = "reset"
	CommandDiscover = "discover"
	CommandPing     = "ping"
	CommandFuncCall = "func_call"
)
