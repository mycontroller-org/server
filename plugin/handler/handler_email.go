package handler

import (
	emailPlugin "github.com/mycontroller-org/server/v2/plugin/handler/email"
)

func init() {
	Register(emailPlugin.PluginEmail, emailPlugin.NewEmailPlugin)
}
