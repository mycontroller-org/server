package virtual_assistant

import (
	vaAlexa "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/assistant/alexa"
	vaGoogle "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/assistant/google"
)

// init plugins
func init() {
	Register(vaGoogle.PluginGoogleAssistant, vaGoogle.New)
	Register(vaAlexa.PluginAlexaAssistant, vaAlexa.New)
}
