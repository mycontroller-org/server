package telemetry

import (
	"context"
	"net/http"
	"time"

	entityAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	httpClient "github.com/mycontroller-org/server/v2/pkg/utils/http_client_json"
	"github.com/mycontroller-org/server/v2/pkg/version"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	gatewayTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

const (
	TELEMETRY_ID  = "MC198501010915"
	TELEMETRY_URL = "https://telemetry.mycontroller.org/event"
	API_VERSION   = "1"
	timeout       = time.Second * 10
)

// reports telemetry data
func ReportTelemetryData(ctx context.Context) {
	logger, err := contextTY.LoggerFromContext(ctx)
	if err != nil {
		// logger not available to report the error
		return
	}
	api, err := entityAPI.FromContext(ctx)
	if err != nil {
		logger.Error("error on getting api", zap.Error(err))
		return
	}

	logger.Debug("collecting telemetry data")

	// create and update version details
	ver := version.Get()
	// update the anonymous id
	telemetryConfig, err := api.Settings().GetTelemetry()
	if err != nil {
		logger.Debug("error on getting telemetry details", zap.Error(err))
		return // if we can't get the anonymous id return from here
	}

	payload := Payload{
		APIVersion:  API_VERSION,
		TelemetryID: TELEMETRY_ID,
		UserID:      telemetryConfig.AnonymousID,
		Application: Application{
			Version:   ver.Version,
			GitCommit: ver.GitCommit,
			BuildDate: ver.BuildDate,
			Platform:  ver.Platform,
			Arch:      ver.Arch,
			GoLang:    ver.GoVersion,
			RunningIn: api.Status().RunningIn(),
			Uptime:    api.Status().Get().Uptime,
			Gateways:  []string{},
			Handlers:  []string{},
		},
	}

	// include city, region and country details
	location, err := api.Settings().GetLocation()
	if err != nil {
		logger.Debug("error on getting location details", zap.Error(err))
	} else {
		payload.Location = Location{
			City:    location.City,
			Region:  location.Region,
			Country: location.Country,
		}
	}

	enabledFilter := []storageTY.Filter{{Key: types.KeyEnabled, Operator: storageTY.OperatorEqual, Value: "true"}}
	pagination := &storageTY.Pagination{Limit: 100, Offset: 0} // gets only the first 100

	// update gateways type in use
	result, err := api.Gateway().List(enabledFilter, pagination)
	if err != nil {
		logger.Error("error on getting gateway details", zap.Error(err))
	} else if result.Count > 0 {
		if data, ok := result.Data.(*[]gatewayTY.Config); ok {
			gateways := make([]string, 0)
			for _, gw := range *data {
				gwType := gw.Provider.GetString("type")
				if !utils.ContainsString(gateways, gwType) {
					gateways = append(gateways, gwType)
				}
			}
			payload.Application.Gateways = gateways
		}
	}

	// update handlers type in use
	result, err = api.Handler().List(enabledFilter, pagination)
	logger.Debug("handler", zap.Any("result", result))
	if err != nil {
		logger.Error("error on getting handler details", zap.Error(err))
	} else if result.Count > 0 {
		if data, ok := result.Data.(*[]handlerTY.Config); ok {
			handlers := make([]string, 0)
			for _, handler := range *data {
				handlerType := handler.Type
				if !utils.ContainsString(handlers, handlerType) {
					handlers = append(handlers, handlerType)
				}
			}
			payload.Application.Handlers = handlers
		}
	}

	logger.Debug("telemetry data to be reported", zap.Any("data", payload))

	// publish the data
	client := httpClient.GetClient(false, timeout)
	response, err := client.ExecuteJson(TELEMETRY_URL, http.MethodPost, nil, nil, payload, http.StatusOK)
	if err != nil {
		if response != nil {
			logger.Debug("error on sending telemetry data", zap.String("response", response.StringBody()), zap.Any("responseConfig", response), zap.Error(err))
		} else {
			logger.Debug("error on sending telemetry data", zap.Error(err))
		}
	}
}
