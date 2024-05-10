package importexport

import (
	"context"

	settings "github.com/mycontroller-org/server/v2/pkg/api/settings"
	encryptionAPI "github.com/mycontroller-org/server/v2/pkg/encryption"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	filterUtils "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	backupRestore "github.com/mycontroller-org/server/v2/plugin/database/storage/backup"
	backupTY "github.com/mycontroller-org/server/v2/plugin/database/storage/backup"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

type BackupAPI struct {
	ctx           context.Context
	logger        *zap.Logger
	backupRestore *backupRestore.BackupRestore
	bus           busTY.Plugin
	settingsAPI   *settings.SettingsAPI
}

func New(ctx context.Context, logger *zap.Logger, backupRestore *backupRestore.BackupRestore, storage storageTY.Plugin, bus busTY.Plugin, enc *encryptionAPI.Encryption) *BackupAPI {
	return &BackupAPI{
		ctx:           ctx,
		logger:        logger.Named("backup_api"),
		backupRestore: backupRestore,
		bus:           bus,
		settingsAPI:   settings.New(ctx, logger, storage, enc, bus),
	}
}

// List by filter and pagination
func (bk *BackupAPI) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	files, err := bk.GetBackupFilesList()
	if err != nil {
		return nil, err
	}

	finalList := make([]interface{}, 0)
	totalCount := int64(0)
	if len(files) > 0 {
		if pagination == nil {
			pagination = &storageTY.Pagination{
				Limit:  10,
				Offset: 0,
				SortBy: []storageTY.Sort{{Field: "id", OrderBy: storageTY.SortByASC}},
			}
		}

		// filter and then sort the files
		filteredFiles := filterUtils.Filter(files, filters, false)
		sortedFiles, count := filterUtils.Sort(filteredFiles, pagination)
		totalCount = count
		finalList = sortedFiles
	}

	result := &storageTY.Result{
		Count:  totalCount,
		Limit:  pagination.Limit,
		Offset: pagination.Offset,
		Data:   finalList,
	}

	return result, nil
}

// Delete backup files
func (bk *BackupAPI) Delete(IDs []string) (int64, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}

	files, err := bk.GetBackupFilesList()
	if err != nil {
		return 0, err
	}

	finalList := make([]interface{}, 0)
	if len(files) > 0 {
		finalList = filterUtils.Filter(files, filters, false)
	}

	deletedCount := int64(0)
	for _, file := range finalList {
		exportedFile, ok := file.(backupTY.BackupFile)
		if !ok {
			continue
		}

		err = utils.RemoveFileOrEmptyDir(exportedFile.FullPath)
		if err != nil {
			return deletedCount, err
		}
		deletedCount++
	}

	return deletedCount, nil
}
