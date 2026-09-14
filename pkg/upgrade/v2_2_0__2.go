package upgrade

import (
	"context"
	"strings"

	entitiesAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	"github.com/mycontroller-org/server/v2/pkg/types"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	svcAccountTY "github.com/mycontroller-org/server/v2/pkg/types/service_account"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

const oldServiceTokenEntity = "service_token"

// Rename storage entity service_token → service_account and rewrite policy
// resource kinds servicetoken → serviceaccount.
func upgrade_2_2_0__2(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, api *entitiesAPI.API) error {
	if err := migrateServiceTokenEntity(logger, storage); err != nil {
		return err
	}
	if err := rewriteServiceTokenPolicyResources(logger, storage); err != nil {
		return err
	}
	if err := api.Policy().EnsureBuiltInPolicies(); err != nil {
		logger.Error("error on refreshing built-in policies", zap.Error(err))
		return err
	}
	return nil
}

func migrateServiceTokenEntity(logger *zap.Logger, storage storageTY.Plugin) error {
	accounts := make([]svcAccountTY.ServiceAccount, 0)
	result, err := storage.Find(oldServiceTokenEntity, &accounts, []storageTY.Filter{}, &storageTY.Pagination{})
	if err != nil {
		logger.Info("no service_token entity to migrate", zap.Error(err))
		return nil
	}
	data, ok := result.Data.(*[]svcAccountTY.ServiceAccount)
	if !ok {
		data = &accounts
	}
	if len(*data) == 0 {
		return nil
	}

	ids := make([]string, 0, len(*data))
	for i := range *data {
		account := (*data)[i]
		filters := []storageTY.Filter{{Key: types.KeyID, Value: account.ID}}
		if err := storage.Upsert(types.EntityServiceAccount, &account, filters); err != nil {
			logger.Error("error on migrating service account", zap.String("id", account.ID), zap.Error(err))
			return err
		}
		ids = append(ids, account.ID)
	}
	if _, err := storage.Delete(oldServiceTokenEntity, []storageTY.Filter{
		{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids},
	}); err != nil {
		logger.Error("error on deleting migrated service_token rows", zap.Error(err))
		return err
	}
	logger.Info("migrated service_token rows to service_account", zap.Int("count", len(ids)))
	return nil
}

func rewriteServiceTokenPolicyResources(logger *zap.Logger, storage storageTY.Plugin) error {
	policies := make([]policyTY.Policy, 0)
	result, err := storage.Find(types.EntityPolicy, &policies, []storageTY.Filter{}, &storageTY.Pagination{})
	if err != nil {
		logger.Info("no policies to rewrite", zap.Error(err))
		return nil
	}
	data, ok := result.Data.(*[]policyTY.Policy)
	if !ok {
		data = &policies
	}

	updated := 0
	for i := range *data {
		policy := (*data)[i]
		changed := false
		for si := range policy.Statements {
			for ri, resource := range policy.Statements[si].Resources {
				rewritten := rewriteServiceTokenResource(resource)
				if rewritten != resource {
					policy.Statements[si].Resources[ri] = rewritten
					changed = true
				}
			}
		}
		if !changed {
			continue
		}
		filters := []storageTY.Filter{{Key: types.KeyID, Value: policy.ID}}
		if err := storage.Upsert(types.EntityPolicy, &policy, filters); err != nil {
			logger.Error("error on rewriting policy resources", zap.String("id", policy.ID), zap.Error(err))
			return err
		}
		updated++
	}
	if updated > 0 {
		logger.Info("rewrote servicetoken policy resources to serviceaccount", zap.Int("count", updated))
	}
	return nil
}

func rewriteServiceTokenResource(resource string) string {
	switch {
	case resource == "servicetoken" || strings.HasPrefix(resource, "servicetoken:"):
		return "serviceaccount" + strings.TrimPrefix(resource, "servicetoken")
	case resource == "service_token" || strings.HasPrefix(resource, "service_token:"):
		return "serviceaccount" + strings.TrimPrefix(resource, "service_token")
	default:
		return resource
	}
}
