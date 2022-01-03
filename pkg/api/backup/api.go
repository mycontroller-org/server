package importexport

import (
	types "github.com/mycontroller-org/server/v2/pkg/types"
	backupTY "github.com/mycontroller-org/server/v2/pkg/types/backup"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	filterUtils "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
)

// List by filter and pagination
func List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	files, err := GetBackupFilesList()
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
		sortedFiles, count := filterUtils.Sort(files, pagination)
		totalCount = count
		finalList = filterUtils.Filter(sortedFiles, filters, false)
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
func Delete(IDs []string) (int64, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}

	files, err := GetBackupFilesList()
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
