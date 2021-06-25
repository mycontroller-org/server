package disk

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/handler"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	helper "github.com/mycontroller-org/backend/v2/pkg/utils/filter_sort"
	yamlUtils "github.com/mycontroller-org/backend/v2/pkg/utils/yaml"
	backupUtil "github.com/mycontroller-org/backend/v2/plugin/handler/backup/util"
	"github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
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
	handlerCfg *handlerML.Config
	cfg        *Config
}

// Init disk backup client
func Init(cfg *handlerML.Config, spec map[string]interface{}) (*Client, error) {
	config := &Config{}
	err := utils.MapToStruct(utils.TagNameNone, spec, config)
	if err != nil {
		return nil, err
	}

	client := &Client{
		handlerCfg: cfg,
		cfg:        config,
	}

	return client, nil
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
func (c *Client) State() *model.State {
	if c.handlerCfg != nil {
		if c.handlerCfg.State == nil {
			c.handlerCfg.State = &model.State{}
		}
		return c.handlerCfg.State
	}
	return &model.State{}
}

// Post func
func (c *Client) Post(data map[string]interface{}) error {
	for name, value := range data {
		zap.L().Debug("processing a request", zap.String("name", name), zap.Any("value", value))
		stringValue, ok := value.(string)
		if !ok {
			continue
		}

		genericData := handlerML.GenericData{}
		err := json.Unmarshal([]byte(stringValue), &genericData)
		if err != nil {
			continue
		}
		if genericData.Type != handlerML.DataTypeBackup {
			continue
		}

		backupConfigData := handlerML.BackupData{}
		err = yamlUtils.UnmarshalBase64Yaml(genericData.Data, &backupConfigData)
		if err != nil {
			zap.L().Error("error on converting backup config data", zap.Error(err), zap.String("name", name), zap.String("value", stringValue))
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

	zap.L().Debug("data", zap.Any("config", newConfig))

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
	zap.L().Debug("Backup job triggered", zap.String("handler", c.handlerCfg.ID))
	// start backup
	filename, err := backupUtil.Backup(prefix, targetExportType)
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

	zap.L().Debug("Export job completed", zap.String("handler", c.handlerCfg.ID), zap.String("timeTaken", time.Since(start).String()))

	err = c.executeRetentionCount(targetDirectory, prefix, targetExportType, retentionCount)
	if err != nil {
		zap.L().Error("error on executing retention count", zap.String("handler", c.handlerCfg.ID), zap.Error(err))
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
	sortBy := []storage.Sort{{Field: "name", OrderBy: storage.SortByDESC}}
	ordered, _ := helper.Sort(matchingFiles, &storage.Pagination{SortBy: sortBy, Limit: -1, Offset: 0})

	if len(ordered) > retentionCount {
		deleteFiles := ordered[retentionCount:]
		for _, f := range deleteFiles {
			if file, ok := f.(model.File); ok {
				zap.L().Debug("deleting a file", zap.Any("file", file))
				filename := fmt.Sprintf("%s/%s", targetDir, file.Name)
				err = utils.RemoveFileOrEmptyDir(filename)
				if err != nil {
					zap.L().Error("error on deleting a file", zap.Any("file", file), zap.Error(err))
				}
			}
		}
	}

	return nil
}
