package systemmonitoring

import (
	"context"
	"fmt"
	"strings"

	"github.com/mycontroller-org/server/v2/pkg/utils"
	metricTY "github.com/mycontroller-org/server/v2/plugin/database/metric/types"
	"github.com/shirou/gopsutil/v4/sensors"
	"go.uber.org/zap"
)

func (p *Provider) updateTemperature() {
	temperatures, err := sensors.TemperaturesWithContext(context.Background())
	if err != nil && !strings.Contains(err.Error(), "Number of warnings") {
		p.logger.Error("error on getting usage", zap.Error(err))
		return
	}

	// presentation message
	presentMsg := p.getSourcePresentationMsg("temperature", "Temperature")

	err = p.postMsg(&presentMsg)
	if err != nil {
		p.logger.Error("error on posting msg", zap.Error(err))
		return
	}

	// data message
	msg := p.getMsg("temperature")

	idList := make([]string, 0)

	for index, temp := range temperatures {
		// convert as unique id
		fieldID := temp.SensorKey
		if utils.ContainsString(idList, fieldID) {
			fieldID = fmt.Sprintf("%s_%v", temp.SensorKey, index)
		}
		idList = append(idList, fieldID)

		// do not include empty sensors
		if temp.Critical == 0 && temp.High == 0 && temp.Temperature == 0 {
			continue
		}

		// if enabled length is greater than 0, do not include other than this list
		if len(p.HostConfig.Temperature.Enabled) > 0 {
			if !utils.ContainsString(p.HostConfig.Temperature.Enabled, fieldID) {
				continue
			}
		}

		// include temperature data
		data := p.getData(fieldID, temp.Temperature, metricTY.MetricTypeGaugeFloat, metricTY.UnitCelsius, true)
		data.Others.Set("high", temp.High, nil)
		data.Others.Set("critical", temp.Critical, nil)
		data.Others.Set("index", index, nil)
		msg.Payloads = append(msg.Payloads, data)
	}

	err = p.postMsg(&msg)
	if err != nil {
		p.logger.Error("error on posting msg", zap.Error(err))
		return
	}
}
