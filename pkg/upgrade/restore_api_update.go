package upgrade

import (
	"context"

	semver "github.com/Masterminds/semver/v3"
	backupTY "github.com/mycontroller-org/server/v2/plugin/database/storage/backup"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

// updates required apis on restore
// this will allow to restore from any lower version
// if there is a schema changed on a version, old schema can be handled with this
// data will be migrated on startup via upgrade scripts
func UpdateStorageRestoreApiMap(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, backupVersion string, apiMap map[string]backupTY.Backup) (map[string]backupTY.Backup, error) {
	backupSemver, err := semver.NewVersion(backupVersion)
	if err != nil {
		logger.Error("error on parsing backup version", zap.String("backupVersion", backupVersion), zap.Error(err))
		return nil, err
	}

	// there is change on version 2.1.1 on "virtual_devices"
	if backupSemver.LessThan(semver.MustParse("2.1.1")) {
		logger.Info("backup is from 2.1.0 or lower version of server, updating required schema changes")
		return updateRestoreApiMap_2_1_1(ctx, logger, storage, apiMap)
	}

	return apiMap, nil
}
