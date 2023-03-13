package topic

// topics used across the application
const (
	TopicInternalShutdown              = "internal.shutdown"                   // request to shutdown the server
	TopicInternalSystemJobs            = "internal.system_jobs"                // system jobs update notification
	TopicPostMessageToProcessor        = "message.to_message_processor"        // message processor, process the received messages from provider
	TopicPostMessageToProvider         = "message.to_provider"                 // provider listens. append gateway id
	TopicPostRawMessageAcknowledgement = "message.raw_message_acknowledgement" // raw message acknowledge
	TopicPostMessageNotifyHandler      = "message.notify_handler"              // post to notify handler
	TopicServiceResourceServer         = "service.resource_server"             // a server listens on this topic, and serves the request
	TopicServiceGateway                = "service.gateway"                     // gateways listen this topic and perform actions like load, reload, stop, start, etc.,
	TopicServiceHandler                = "service.handler"                     // handler listen this topic and perform actions like load, reload, stop, start, etc.,
	TopicServiceTask                   = "service.task"                        // tasks listen this topic and perform actions like load, reload, stop, start, etc.,
	TopicServiceScheduler              = "service.scheduler"                   // scheduler listen this topic and perform actions like add, remove, disable, etc.,
	TopicServiceVirtualAssistant       = "service.virtual_assistant"           // virtual assistant listen this topic and perform actions like add, remove, disable, etc.,
	TopicEventsAll                     = "event.>"                             // all events
	TopicEventGateway                  = "event.gateway"                       // gateway events
	TopicEventNode                     = "event.node"                          // node events
	TopicEventSource                   = "event.source"                        // source events
	TopicEventField                    = "event.field"                         // field events
	TopicEventTask                     = "event.task"                          // task events
	TopicEventSchedule                 = "event.schedule"                      // schedule events
	TopicEventHandler                  = "event.handler"                       // handler events
	TopicEventFirmware                 = "event.firmware"                      // firmware create/update/delete events
	TopicEventDataRepository           = "event.data_repository"               // data repository events
	TopicEventForwardPayload           = "event.forward_payload"               // forward payload events
	TopicEventVirtualDevice            = "event.virtual_device"                // virtual device events
	TopicEventVirtualAssistant         = "event.virtual_assistant"             // virtual assistant events
	TopicFirmwareBlocks                = "firmware.blocks"                     // request to shutdown the server
)
