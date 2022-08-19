package alexa

import (
	"github.com/mycontroller-org/server/v2/pkg/utils"
	alexaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/assistant/alexa/types"
)

func getErrorResponse(endpointID, errorType, message string) *alexaTY.Response {
	return &alexaTY.Response{
		Event: alexaTY.DirectiveOrEvent{
			Header: alexaTY.Header{
				Namespace:      "Alexa",
				Name:           "ErrorResponse",
				MessageID:      utils.RandUUID(),
				PayloadVersion: "3",
			},
			Endpoint: &alexaTY.DirectiveEndpoint{
				EndpointID: endpointID,
			},
			Payload: map[string]interface{}{
				"type":    errorType,
				"message": message,
			},
		},
	}
}
