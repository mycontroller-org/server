package disk

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types"
	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	helper "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	yamlUtils "github.com/mycontroller-org/server/v2/pkg/utils/yaml"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	backupUtil "github.com/mycontroller-org/server/v2/plugin/handler/backup/util"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

const (
	PluginBackupDisk = "backup_disk"
	loggerName       = "handler_backup_disk"
)

// Config of disk backup client
type Config struct {
	Prefix            string
	StorageExportType string
	TargetDirectory   string
	RetentionCount    int
}

// Client struct
type Client struct {
	ctx        context.Context
	handlerCfg *handlerTY.Config
	cfg        *Config
	logger     *zap.Logger
	storage    storageTY.Plugin
	bus        busTY.Plugin
}

// New disk backup client
func New(ctx context.Context, cfg *handlerTY.Config, spec map[string]interface{}) (*Client, error) {
	logger, err := contextTY.LoggerFromContext(ctx)
	if err != nil {
		return nil, err
	}
	storage, err := storageTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	bus, err := busTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = utils.MapToStruct(utils.TagNameNone, spec, config)
	if err != nil {
		return nil, err
	}

	client := &Client{
		ctx:        ctx,
		handlerCfg: cfg,
		cfg:        config,
		logger:     logger.Named(loggerName),
		storage:    storage,
		bus:        bus,
	}

	return client, nil
}

func (p *Client) Name() string {
	return PluginBackupDisk
}

// Start func
func (c *Client) Start() error {
	return nil
}

// Close Func
func (c *Client) Close() error {
	return nil
}

// State func
func (c *Client) State() *types.State {
	if c.handlerCfg != nil {
		if c.handlerCfg.State == nil {
			c.handlerCfg.State = &types.State{}
		}
		return c.handlerCfg.State
	}
	return &types.State{}
}

// Post func
func (c *Client) Post(data map[string]interface{}) error {
	for name, value := range data {
		c.logger.Debug("processing a request", zap.String("name", name), zap.Any("value", value))
		stringValue, ok := value.(string)
		if !ok {
			continue
		}

		genericData := handlerTY.GenericData{}
		err := json.Unmarshal([]byte(stringValue), &genericData)
		if err != nil {
			continue
		}
		if genericData.Type != handlerTY.DataTypeBackup {
			continue
		}

		backupConfigData := handlerTY.BackupData{}
		err = yamlUtils.UnmarshalBase64Yaml(genericData.Data, &backupConfigData)
		if err != nil {
			c.logger.Error("error on converting backup config data", zap.Error(err), zap.String("name", name), zap.String("value", stringValue))
			continue
		}

		if backupConfigData.ProviderType != backupUtil.ProviderDisk {
			continue
		}

		err = c.triggerBackup(backupConfigData.Spec)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) triggerBackup(spec map[string]interface{}) error {
	newConfig := &Config{}
	err := utils.MapToStruct(utils.TagNameNone, spec, newConfig)
	if err != nil {
		return err
	}

	c.logger.Debug("data", zap.Any("config", newConfig))

	targetExportType := c.cfg.StorageExportType
	targetDirectory := c.cfg.TargetDirectory
	prefix := c.cfg.Prefix
	retentionCount := c.cfg.RetentionCount

	if newConfig.StorageExportType != "" {
		targetExportType = newConfig.StorageExportType
	}

	if newConfig.TargetDirectory != "" {
		targetDirectory = newConfig.TargetDirectory
	}

	targetDirectory = strings.TrimSuffix(targetDirectory, "/")

	if newConfig.Prefix != "" {
		prefix = newConfig.Prefix
	}

	if prefix == "" {
		prefix = c.handlerCfg.ID
	}

	if newConfig.RetentionCount != 0 {
		retentionCount = newConfig.RetentionCount
	}

	start := time.Now()
	c.logger.Debug("Backup job triggered", zap.String("handler", c.handlerCfg.ID))
	// start backup
	filename, err := backupUtil.Backup(c.ctx, c.logger, prefix, targetExportType, c.storage, c.bus)
	if err != nil {
		return err
	}

	// move the file to target location
	// get final file name
	zipFilename := filepath.Base(filename)
	targetLocation := fmt.Sprintf("%s/%s", targetDirectory, zipFilename)
	err = utils.CopyFile(filename, targetLocation)
	if err != nil {
		return err
	}
	err = utils.RemoveFileOrEmptyDir(filename)
	if err != nil {
		return err
	}

	c.logger.Debug("Export job completed", zap.String("handler", c.handlerCfg.ID), zap.String("timeTaken", time.Since(start).String()))

	err = c.executeRetentionCount(targetDirectory, prefix, targetExportType, retentionCount)
	if err != nil {
		c.logger.Error("error on executing retention count", zap.String("handler", c.handlerCfg.ID), zap.Error(err))
	}

	return nil
}

func (c *Client) executeRetentionCount(targetDir, prefix, targetExportType string, retentionCount int) error {
	if retentionCount <= 0 {
		return nil
	}

	files, err := utils.ListFiles(targetDir)
	if err != nil {
		return err
	}

	prefix = fmt.Sprintf("%s_%s_%s", prefix, backupUtil.BackupIdentifier, targetExportType)
	matchingFiles := make([]interface{}, 0)
	for _, file := range files {
		if strings.HasPrefix(file.Name, prefix) {
			matchingFiles = append(matchingFiles, file)
		}
	}

	// sort by filename
	sortBy := []storageTY.Sort{{Field: "name", OrderBy: storageTY.SortByDESC}}
	ordered, _ := helper.Sort(matchingFiles, &storageTY.Pagination{SortBy: sortBy, Limit: -1, Offset: 0})

	if len(ordered) > retentionCount {
		deleteFiles := ordered[retentionCount:]
		for _, f := range deleteFiles {
			if file, ok := f.(types.File); ok {
				c.logger.Debug("deleting a file", zap.Any("file", file))
				filename := fmt.Sprintf("%s/%s", targetDir, file.Name)
				err = utils.RemoveFileOrEmptyDir(filename)
				if err != nil {
					c.logger.Error("error on deleting a file", zap.Any("file", file), zap.Error(err))
				}
			}
		}
	}

	return nil
}
