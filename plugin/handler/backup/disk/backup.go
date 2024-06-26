package disk

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	helper "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
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
	Disabled             string // used globally
	Type                 string // used globally
	ProviderType         string // used globally
	Prefix               string
	StorageExportType    string
	TargetDirectory      string
	RetentionCount       int
	IncludeSecureShare   bool // include secure directory on the backup
	IncludeInsecureShare bool // include insecure directory on the backup
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

// disk backup client
func New(ctx context.Context, cfg *handlerTY.Config) (*Client, error) {
	logger, err := loggerUtils.FromContext(ctx)
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
	err = utils.MapToStruct(utils.TagNameNone, cfg.Spec, config)
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
func (c *Client) Post(parameters map[string]interface{}) error {
	for name, rawParameter := range parameters {
		parameter, ok := handlerTY.IsTypeOf(rawParameter, handlerTY.DataTypeBackup)
		if !ok {
			continue
		}
		c.logger.Debug("data", zap.Any("name", name), zap.Any("parameter", parameter))

		backupConfigData := handlerTY.BackupData{}
		err := utils.MapToStruct(utils.TagNameNone, parameter, &backupConfigData)
		if err != nil {
			c.logger.Error("error on converting backup config data", zap.Error(err), zap.String("name", name), zap.Any("parameter", parameter))
			continue
		}

		if backupConfigData.ProviderType != backupUtil.ProviderDisk {
			continue
		}

		err = c.triggerBackup(parameter)
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

	// get base directory for storage
	baseDir := types.GetEnvString(types.ENV_DIR_DATA_STORAGE)
	if baseDir == "" {
		return fmt.Errorf("environment '%s' not set", types.ENV_DIR_DATA_STORAGE)
	}

	// start backup
	filename, err := backupUtil.Backup(c.ctx, c.logger, baseDir, prefix, targetExportType, c.cfg.IncludeSecureShare, c.cfg.IncludeInsecureShare, c.storage, c.bus)
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
