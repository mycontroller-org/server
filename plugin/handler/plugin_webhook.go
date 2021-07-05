package handler

import (
	webhook "github.com/mycontroller-org/server/v2/plugin/handler/webhook"
)

func init() {
	Register(webhook.PluginWebhook, webhook.NewWebhookPlugin)
}
