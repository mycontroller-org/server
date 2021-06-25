package resource

import (
	"fmt"
	"strings"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/model"
	handlerML "github.com/mycontroller-org/server/v2/pkg/model/handler"
	rsML "github.com/mycontroller-org/server/v2/pkg/model/resource_service"
	coreScheduler "github.com/mycontroller-org/server/v2/pkg/service/core_scheduler"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	yamlUtils "github.com/mycontroller-org/server/v2/pkg/utils/yaml"
	"go.uber.org/zap"
)

// Client struct
type Client struct {
	HandlerCfg *handlerML.Config
}

const (
	schedulePrefix = "resource_handler"
)

// Start handler implementation
func (c *Client) Start() error { return nil }

// Close handler implementation
func (c *Client) Close() error {
	c.unloadAll()
	return nil
}

// State implementation
func (c *Client) State() *model.State {
	if c.HandlerCfg != nil {
		if c.HandlerCfg.State == nil {
			c.HandlerCfg.State = &model.State{}
		}
		return c.HandlerCfg.State
	}
	return &model.State{}
}

// Post handler implementation
func (c *Client) Post(data map[string]interface{}) error {
	for name, value := range data {
		stringValue, ok := value.(string)
		if !ok {
			continue
		}

		genericData := handlerML.GenericData{}
		err := json.Unmarshal([]byte(stringValue), &genericData)
		if err != nil {
			continue
		}

		if !strings.HasPrefix(genericData.Type, handlerML.DataTypeResource) {
			continue
		}

		rsData := handlerML.ResourceData{}
		err = yamlUtils.UnmarshalBase64Yaml(genericData.Data, &rsData)
		if err != nil {
			zap.L().Error("error on loading resource data", zap.Error(err), zap.String("name", name), zap.String("input", stringValue))
			continue
		}

		if rsData.PreDelay != "" {
			delayDuration, err := time.ParseDuration(rsData.PreDelay)
			if err != nil {
				return fmt.Errorf("invalid preDelay. name:%s, preDelay:%s", name, rsData.PreDelay)
			}
			if delayDuration > 0 {
				c.schedule(name, rsData)
				continue
			}
		}

		zap.L().Debug("about to perform an action", zap.String("rawData", stringValue), zap.Any("finalData", rsData))
		busUtils.PostToResourceService("resource_fake_id", rsData, rsML.TypeResourceAction, rsML.CommandSet, "")
	}
	return nil
}

// preDelay scheduler helpers

func (c *Client) getScheduleTriggerFunc(name string, rsData handlerML.ResourceData) func() {
	return func() {
		// disable the schedule
		c.unschedule(name)

		// call the resource action
		zap.L().Debug("scheduler triggered. about to perform an action", zap.String("name", name), zap.Any("rsData", rsData))
		busUtils.PostToResourceService("resource_fake_id", rsData, rsML.TypeResourceAction, rsML.CommandSet, "")
	}
}

func (c *Client) schedule(name string, rsData handlerML.ResourceData) {
	c.unschedule(name) // removes the existing schedule, if any
	schedulerID := c.getScheduleID(name)
	cronSpec := fmt.Sprintf("@every %s", rsData.PreDelay)
	err := coreScheduler.SVC.AddFunc(schedulerID, cronSpec, c.getScheduleTriggerFunc(name, rsData))
	if err != nil {
		zap.L().Error("error on adding schedule", zap.Error(err))
	}
	zap.L().Debug("added a schedule", zap.String("name", name), zap.String("schedulerID", schedulerID), zap.Any("resourceData", rsData))
}

func (c *Client) unschedule(name string) {
	schedulerID := c.getScheduleID(name)
	coreScheduler.SVC.RemoveFunc(schedulerID)
	zap.L().Debug("removed a schedule", zap.String("name", name), zap.String("schedulerID", schedulerID))
}

func (c *Client) unloadAll() {
	coreScheduler.SVC.RemoveWithPrefix(fmt.Sprintf("%s_%s", schedulePrefix, c.HandlerCfg.ID))
}

func (c *Client) getScheduleID(name string) string {
	return fmt.Sprintf("%s_%s_%s", schedulePrefix, c.HandlerCfg.ID, name)
}
