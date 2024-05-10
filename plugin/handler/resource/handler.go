package resource

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

const (
	PluginResourceHandler = "resource"

	schedulePrefix = "resource_handler"
	loggerName     = "handler_resource"
)

// ResourceClient struct
type ResourceClient struct {
	HandlerCfg *handlerTY.Config
	store      *store
	logger     *zap.Logger
	scheduler  schedulerTY.CoreScheduler
	bus        busTY.Plugin
}

func NewResourcePlugin(ctx context.Context, config *handlerTY.Config) (handlerTY.Plugin, error) {
	logger, err := loggerUtils.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	scheduler, err := schedulerTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	bus, err := busTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	return &ResourceClient{
		HandlerCfg: config,
		store:      &store{mutex: sync.RWMutex{}, handlerID: config.ID, jobs: map[string]JobsConfig{}},
		logger:     logger.Named(loggerName),
		scheduler:  scheduler,
		bus:        bus,
	}, nil
}

func (p *ResourceClient) Name() string {
	return PluginResourceHandler
}

// Start handler implementation
func (c *ResourceClient) Start() error {
	// load handler data from disk
	err := c.store.loadFromDisk(c)
	if err != nil {
		c.logger.Error("failed to load handler data", zap.String("diskLocation", c.store.getName()), zap.Error(err))
		return nil
	}

	return nil
}

// Close handler implementation
func (c *ResourceClient) Close() error {
	c.unloadAll()

	// save jobs to disk location
	err := c.store.saveToDisk()
	if err != nil {
		c.logger.Error("failed to save handler data", zap.String("diskLocation", c.store.getName()), zap.Error(err))
	}
	return nil
}

// State implementation
func (c *ResourceClient) State() *types.State {
	if c.HandlerCfg != nil {
		if c.HandlerCfg.State == nil {
			c.HandlerCfg.State = &types.State{}
		}
		return c.HandlerCfg.State
	}
	return &types.State{}
}

// Post handler implementation
func (c *ResourceClient) Post(parameters map[string]interface{}) error {
	for name := range parameters {
		rawParameter := parameters[name]
		parameter, ok := handlerTY.HasTypePrefixOf(rawParameter, handlerTY.DataTypeResource)
		if !ok {
			continue
		}
		c.logger.Debug("data", zap.Any("name", name), zap.Any("parameter", parameter))

		rsData := handlerTY.ResourceData{}
		err := utils.MapToStruct(utils.TagNameNone, parameter, &rsData)
		if err != nil {
			c.logger.Error("error on loading resource data", zap.Error(err), zap.String("name", name), zap.Any("input", parameter))
			continue
		}

		if rsData.PreDelay != "" {
			delayDuration, err := time.ParseDuration(rsData.PreDelay)
			if err != nil {
				return fmt.Errorf("invalid preDelay. name:%s, preDelay:%s", name, rsData.PreDelay)
			}
			if delayDuration > 0 {
				c.store.add(name, rsData)
				c.schedule(name, rsData)
				continue
			}
		}

		c.logger.Debug("about to perform an action", zap.Any("rawData", parameter), zap.Any("finalData", rsData))
		busUtils.PostToResourceService(c.logger, c.bus, "resource_fake_id", rsData, rsTY.TypeResourceAction, rsTY.CommandSet, "")
	}
	return nil
}

// preDelay scheduler helpers

func (c *ResourceClient) getScheduleTriggerFunc(name string, rsData handlerTY.ResourceData) func() {
	return func() {
		// disable the schedule
		c.unschedule(name)

		// call the resource action
		c.logger.Debug("scheduler triggered. about to perform an action", zap.String("name", name), zap.Any("rsData", rsData))
		busUtils.PostToResourceService(c.logger, c.bus, "resource_fake_id", rsData, rsTY.TypeResourceAction, rsTY.CommandSet, "")
	}
}

func (c *ResourceClient) schedule(name string, rsData handlerTY.ResourceData) {
	c.unschedule(name) // removes the existing schedule, if any

	schedulerID := c.getScheduleID(name)
	cronSpec := fmt.Sprintf("@every %s", rsData.PreDelay)
	err := c.scheduler.AddFunc(schedulerID, cronSpec, c.getScheduleTriggerFunc(name, rsData))
	if err != nil {
		c.logger.Error("error on adding schedule", zap.Error(err))
	}
	c.logger.Debug("added a schedule", zap.String("name", name), zap.String("schedulerID", schedulerID), zap.Any("resourceData", rsData))
}

func (c *ResourceClient) unschedule(name string) {
	schedulerID := c.getScheduleID(name)
	c.scheduler.RemoveFunc(schedulerID)
	c.logger.Debug("removed a schedule", zap.String("name", name), zap.String("schedulerID", schedulerID))
}

func (c *ResourceClient) unloadAll() {
	c.scheduler.RemoveWithPrefix(fmt.Sprintf("%s_%s", schedulePrefix, c.HandlerCfg.ID))
}

func (c *ResourceClient) getScheduleID(name string) string {
	return fmt.Sprintf("%s_%s_%s", schedulePrefix, c.HandlerCfg.ID, name)
}
