package handler

import (
	emailPlugin "github.com/mycontroller-org/server/v2/plugin/handler/email"
	resource "github.com/mycontroller-org/server/v2/plugin/handler/resource"
	telegram "github.com/mycontroller-org/server/v2/plugin/handler/telegram"
	webhook "github.com/mycontroller-org/server/v2/plugin/handler/webhook"
)

func init() {
	Register(emailPlugin.PluginEmail, emailPlugin.NewEmailPlugin)
	Register(resource.PluginResourceHandler, resource.NewResourcePlugin)
	Register(telegram.PluginTelegram, telegram.NewTelegramPlugin)
	Register(webhook.PluginWebhook, webhook.NewWebhookPlugin)
}
