package upgrade

import (
	"context"
	"sort"

	semver "github.com/Masterminds/semver/v3"
	entitiesAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

var (
	semanticVersions []*semver.Version
)

func init() {
	// parse parseSemVer
	semvers, err := versions(upgrades)
	if err != nil {
		panic(err)
	}
	semanticVersions = semvers
}

// version return the list of semantic version sorted
func versions(versions map[string]upgradeFunction) ([]*semver.Version, error) {
	versionLists := make([]*semver.Version, len(versions))
	versionIndex := 0
	for v := range versions {
		semv, err := semver.NewVersion(v)
		if err != nil {
			return nil, err
		}
		versionLists[versionIndex] = semv
		versionIndex++
	}

	// apply the updates in order
	sort.Sort(semver.Collection(versionLists))
	return versionLists, nil
}

// returns the most recent patch version
// will be used in on new installation, add in the database as a reference,
// to avoid upgrades on fresh installation
func GetLatestUpgradeVersion() string {
	size := len(semanticVersions)
	return semanticVersions[size-1].String()
}

// starts the upgrade process
func StartUpgrade(ctx context.Context, currentVersion string) (string, error) {
	_logger, err := contextTY.LoggerFromContext(ctx)
	if err != nil {
		return "", err
	}
	logger := _logger.Named("upgrade")

	appliedPatch := ""
	currentSemVersion, err := semver.NewVersion(currentVersion)
	if err != nil {
		logger.Error("error on parsing version", zap.String("input", currentVersion), zap.Error(err))
		return appliedPatch, err
	}

	storage, err := storageTY.FromContext(ctx)
	if err != nil {
		logger.Error("error on getting storage", zap.Error(err))
		return appliedPatch, err
	}

	api, err := entitiesAPI.FromContext(ctx)
	if err != nil {
		logger.Error("error on getting entities api", zap.Error(err))
		return appliedPatch, err
	}

	// get all the required components
	for _, v := range semanticVersions {
		if v.GreaterThan(currentSemVersion) {
			logger.Info("upgrade triggered", zap.String("patchVersion", v.String()))
			err = upgrades[v.String()](ctx, logger, storage, api)
			if err != nil {
				logger.Error("error on upgrade", zap.Error(err))
				return appliedPatch, err
			}
			appliedPatch = v.String()

			// update applied patch details into database
			sysVersion, err := api.Settings().GetVersion()
			if err != nil {
				logger.Error("error on getting version details from database", zap.Error(err))
				return appliedPatch, err
			}

			// update last applied patch
			sysVersion.LastUpgrade = appliedPatch
			err = api.Settings().UpdateVersion(sysVersion)
			if err != nil {
				logger.Error("error on setting version details to database", zap.Error(err))
				return appliedPatch, err
			}
		}
	}
	return appliedPatch, nil
}
