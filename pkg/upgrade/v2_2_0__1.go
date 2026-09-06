package upgrade

import (
	"context"

	entitiesAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

// RBAC introduction: create built-in policies and grant admin to existing users
// that have no policies attached (pre-RBAC installs).
func upgrade_2_2_0__1(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, api *entitiesAPI.API) error {
	policyAPI := api.Policy()

	if err := policyAPI.EnsureBuiltInPolicies(); err != nil {
		logger.Error("error on creating built-in policies", zap.Error(err))
		return err
	}
	logger.Info("built-in access policies ensured (admin, readwrite, readonly)")

	if err := policyAPI.AssignAdminToUsersWithoutPolicies(); err != nil {
		logger.Error("error on assigning admin policy to existing users", zap.Error(err))
		return err
	}
	logger.Info("existing users without policies assigned admin policy")

	return nil
}
