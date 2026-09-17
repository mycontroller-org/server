package upgrade

import (
	"context"

	semver "github.com/Masterminds/semver/v3"
	"github.com/mycontroller-org/server/v2/pkg/types"
	backupTY "github.com/mycontroller-org/server/v2/plugin/database/storage/backup"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

// updates required apis on restore
// this will allow to restore from any lower version
// if there is a schema changed on a version, old schema can be handled with this
// data will be migrated on startup via upgrade scripts
func UpdateStorageRestoreApiMap(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, backupVersion, lastUpgrade string, apiMap map[string]backupTY.Backup) (map[string]backupTY.Backup, error) {
	backupSemver, err := semver.NewVersion(backupVersion)
	if err != nil {
		logger.Error("error on parsing backup version", zap.String("backupVersion", backupVersion), zap.Error(err))
		return nil, err
	}
	// remove pre release (2.1.1-devel => 2.1.1)
	updatedBackupSemver, err := backupSemver.SetPrerelease("")
	if err != nil {
		logger.Error("error on removing preRelease", zap.String("backupVersion", backupVersion), zap.Error(err))
		return nil, err
	}
	logger.Debug("removed preRelease from version details", zap.String("updatedBackupVersion", updatedBackupSemver.Original()), zap.String("lastUpgrade", lastUpgrade))

	// there is change on version 2.1.1 on "virtual_devices"
	if updatedBackupSemver.LessThan(semver.MustParse("2.1.1")) {
		logger.Info("backup is from 2.1.0 or lower version of server, updating required schema changes")
		var err error
		apiMap, err = updateRestoreApiMap_2_1_1(ctx, logger, storage, apiMap)
		if err != nil {
			return nil, err
		}
	}

	// user / service account disabled → enabled
	// Use LastUpgrade: 2.3.0-devel is not older than 2.3.0, but LastUpgrade 2.2.0-2 still needs this.
	if needsUserEnabledRestore(logger, backupVersion, lastUpgrade) {
		logger.Info("user and service account still use disabled, loading old schema")
		var err error
		apiMap, err = updateRestoreApiMap_2_3_0(ctx, logger, storage, apiMap)
		if err != nil {
			return nil, err
		}
	}

	// service_token was renamed to service_account
	if api, ok := apiMap[types.EntityServiceAccount]; ok {
		apiMap[oldServiceTokenEntity] = api
	}

	return apiMap, nil
}

func needsUserEnabledRestore(logger *zap.Logger, backupVersion, lastUpgrade string) bool {
	if lastUpgrade != "" {
		current, err := semver.NewVersion(lastUpgrade)
		if err != nil {
			logger.Error("error on parsing lastUpgrade", zap.String("lastUpgrade", lastUpgrade), zap.Error(err))
			return false
		}
		return current.LessThan(semver.MustParse("2.3.0-1"))
	}
	v, err := semver.NewVersion(backupVersion)
	if err != nil {
		return false
	}
	stripped, err := v.SetPrerelease("")
	if err != nil {
		return false
	}
	return stripped.LessThan(semver.MustParse("2.3.0"))
}
