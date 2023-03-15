package upgrade

import (
	"context"

	entitiesAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

type upgradeFunction = func(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, api *entitiesAPI.API) error

// map all the upgrade paths
// version should start with next release version and maintain sequence of "-1", "-2", to support multiple upgrades
// with this version format also it is possible to upgrade between development version and production version
var upgrades = map[string]upgradeFunction{
	"2.0.0-1": upgrade_2_0_0__1, // 2.0.0 upgrade #1
}
