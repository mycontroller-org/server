package types

import "github.com/mycontroller-org/server/v2/pkg/types/cmap"

// https://developer.amazon.com/en-US/docs/alexa/device-apis/alexa-discovery-objects.html#endpoint-object
type Endpoint struct {
	EndpointID           string                `json:"endpointId"`
	ManufacturerName     string                `json:"manufacturerName"`
	Description          string                `json:"description"`
	FriendlyName         string                `json:"friendlyName"`
	DisplayCategories    []string              `json:"displayCategories"`
	AdditionalAttributes *AdditionalAttributes `json:"additionalAttributes,omitempty"`
	Capabilities         []Capability          `json:"capabilities"`
	Connections          []Connection          `json:"connections,omitempty"`
	Relationships        *Relationships        `json:"relationships,omitempty"`
	Cookie               cmap.CustomStringMap  `json:"cookie"`
}

// https://developer.amazon.com/en-US/docs/alexa/device-apis/alexa-discovery-objects.html#additionalattributes-object
type AdditionalAttributes struct {
	Manufacturer     string `json:"manufacturer,omitempty"`
	Model            string `json:"model,omitempty"`
	SerialNumber     string `json:"serialNumber,omitempty"`
	FirmwareVersion  string `json:"firmwareVersion,omitempty"`
	SoftwareVersion  string `json:"softwareVersion,omitempty"`
	CustomIdentifier string `json:"customIdentifier,omitempty"`
}

// https://developer.amazon.com/en-US/docs/alexa/device-apis/alexa-discovery-objects.html#capability-object
type Capability struct {
	Type                    string                   `json:"type"`
	Interface               string                   `json:"interface"`
	Instance                string                   `json:"instance,omitempty"`
	Version                 string                   `json:"version"`
	Properties              *Properties              `json:"properties,omitempty"`
	CapabilityResources     *cmap.CustomMap          `json:"capabilityResources,omitempty"`
	Configuration           *cmap.CustomMap          `json:"configuration,omitempty"` // or configurations
	Semantics               *Semantics               `json:"semantics,omitempty"`
	VerificationsRequired   []VerificationsRequired  `json:"verificationsRequired,omitempty"`
	DirectiveConfigurations []DirectiveConfiguration `json:"directiveConfigurations,omitempty"`
}

// https://developer.amazon.com/en-US/docs/alexa/device-apis/alexa-discovery-objects.html#directiveconfigurations-object-details
type DirectiveConfiguration struct {
	Directives                             []string                                `json:"directives"`
	RequestedAuthenticationConfidenceLevel *RequestedAuthenticationConfidenceLevel `json:"requestedAuthenticationConfidenceLevel,omitempty"`
}

type RequestedAuthenticationConfidenceLevel struct {
	Level        int           `json:"level"` // Valid values: 400, 500
	CustomPolicy *CustomPolicy `json:"customPolicy,omitempty"`
}

type CustomPolicy struct {
	PolicyName string `json:"policyName,omitempty"` // Valid values: OTP, and VOICE_PIN.
}

// https://developer.amazon.com/en-US/docs/alexa/device-apis/alexa-discovery-objects.html#properties-object
type Properties struct {
	Supported           []cmap.CustomStringMap `json:"supported,omitempty"`
	ProactivelyReported bool                   `json:"proactivelyReported"`
	Retrievable         bool                   `json:"retrievable"`
}

// https://developer.amazon.com/en-US/docs/alexa/device-apis/alexa-discovery-objects.html#connections-object
type Connection struct {
	Type                string `json:"type"` // Valid values: MATTER, TCP_IP,ZIGBEE, ZWAVE, or UNKNOWN
	MacAddress          string `json:"macAddress,omitempty"`
	HomeID              string `json:"homeId,omitempty"`
	NodeID              string `json:"nodeId,omitempty"`
	MatterDiscriminator uint   `json:"matterDiscriminator"` // Integer between 0-4095
	MatterVendorID      uint   `json:"matterVendorId"`      // Integer between 0-65535
	MatterProductID     uint   `json:"matterProductId"`     // Integer between 0-65535
	MacNetworkInterface string `json:"macNetworkInterface,omitempty"`
	Value               string `json:"value,omitempty"` // required when type = UNKNOWN
}

// https://developer.amazon.com/en-US/docs/alexa/device-apis/alexa-discovery-objects.html#relationships-object
type Relationships struct {
	IsConnectedBy *IsConnectedBy `json:"isConnectedBy,omitempty"`
	IsPartOf      *IsPartOf      `json:"Alexa.Automotive.IsPartOf,omitempty"`
}

type IsConnectedBy struct {
	EndpointId string `json:"endpointId"`
}

type IsPartOf struct {
	EndpointId string `json:"endpointId"`
}

// https://developer.amazon.com/en-US/docs/alexa/device-apis/alexa-discovery-objects.html#verifications-required-object
type VerificationsRequired struct {
	Directive string               `json:"directive"`
	Methods   cmap.CustomStringMap `json:"methods"`
}

// https://developer.amazon.com/en-US/docs/alexa/device-apis/alexa-discovery-objects.html#semantics-object
type Semantics struct {
	ActionMappings []ActionMapping `json:"actionMappings,omitempty"`
	StateMappings  []StateMapping  `json:"stateMappings,omitempty"`
}

// https://developer.amazon.com/en-US/docs/alexa/device-apis/alexa-discovery-objects.html#action-mapping
type ActionMapping struct {
	Type      string            `json:"@type"`   // valid value is ActionsToDirective
	Actions   []string          `json:"actions"` // Valid values: Alexa.Actions.Open, Alexa.Actions.Close, Alexa.Actions.Raise, Alexa.Actions.Lower, Alexa.Actions.SetEcoOn, and Alexa.Actions.SetEcoOff.
	Directive EndpointDirective `json:"directive"`
}

type EndpointDirective struct {
	Name    string          `json:"name"`
	Payload *cmap.CustomMap `json:"payload,omitempty"`
}

// https://developer.amazon.com/en-US/docs/alexa/device-apis/alexa-discovery-objects.html#state-mapping
type StateMapping struct {
	Type   string                 `json:"@type"`           // Valid values: one of StatesToValue or StatesToRange.
	States []string               `json:"states"`          // Array of Alexa states that are mapped to your controller values. Valid values: Alexa.States.Open, Alexa.States.Closed, Alexa.States.EcoOn, Alexa.States.EcoOff, Alexa.States.Low, Alexa.States.Empty, Alexa.States.Full, Alexa.States.Done, and Alexa.States.Stuck.
	Value  interface{}            `json:"value,omitempty"` // required when @type is StatesToValue
	Range  map[string]interface{} `json:"range,omitempty"` // required when @type is StatesToRange
}
