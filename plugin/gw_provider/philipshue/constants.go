package philipshue

import (
	"fmt"
	"net/http"

	httpProtocol "github.com/mycontroller-org/backend/v2/plugin/gw_protocol/protocol_http"
)

// global constants
const (
	KeyUsername = "username"
)

type apiConf struct {
	api          string
	method       string
	responseCode int
}

func (ac *apiConf) getRequestConf(id string) *httpProtocol.RequestConfig {
	path := ac.api
	if id != "" {
		path = fmt.Sprintf(ac.api, id)
	}
	return &httpProtocol.RequestConfig{
		Method:       ac.method,
		Path:         path,
		ResponseCode: ac.responseCode,
	}
}

// https://developers.meethue.com/develop/hue-api/
var (
	// lights api
	apiLightsListAll          = apiConf{api: "lights", method: http.MethodGet, responseCode: 200}
	apiLightsListNew          = apiConf{api: "lights/new", method: http.MethodGet, responseCode: 200}
	apiLightsGet              = apiConf{api: "lights/%s", method: http.MethodGet, responseCode: 200}
	apiLightsUpdateAttributes = apiConf{api: "lights/%s", method: http.MethodPut, responseCode: 200}
	apiLightsUpdateState      = apiConf{api: "lights/%s/state", method: http.MethodPut, responseCode: 200}

	// sensors api
	apiSensorsListAll          = apiConf{api: "sensors", method: http.MethodGet, responseCode: 200}
	apiSensorsListNew          = apiConf{api: "sensors/new", method: http.MethodGet, responseCode: 200}
	apiSensorsGet              = apiConf{api: "sensors/%s", method: http.MethodGet, responseCode: 200}
	apiSensorsUpdateAttributes = apiConf{api: "sensors/%s", method: http.MethodPut, responseCode: 200}
	apiSensorsUpdateConfig     = apiConf{api: "sensors/%s/config", method: http.MethodPut, responseCode: 200}
	apiSensorsUpdateState      = apiConf{api: "sensors/%s/state", method: http.MethodPut, responseCode: 200}
)
