package handler

import (
	telegram "github.com/mycontroller-org/server/v2/plugin/handler/telegram"
)

func init() {
	Register(telegram.PluginTelegram, telegram.NewTelegramPlugin)
}
