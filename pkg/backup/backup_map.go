package backup

import (
	"context"
	"fmt"

	entitiesAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/config"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	backupPlugin "github.com/mycontroller-org/server/v2/plugin/database/storage/backup"
	backupTY "github.com/mycontroller-org/server/v2/plugin/database/storage/backup"
	"go.uber.org/zap"
)

// returns import and export entity map
func GetStorageApiMap(ctx context.Context) (map[string]backupTY.Backup, error) {
	entities, err := entitiesAPI.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	funcMap := map[string]backupTY.Backup{
		types.EntityDashboard:        entities.Dashboard(),
		types.EntityDataRepository:   entities.DataRepository(),
		types.EntityField:            entities.Field(),
		types.EntityFirmware:         entities.Firmware(),
		types.EntityForwardPayload:   entities.ForwardPayload(),
		types.EntityGateway:          entities.Gateway(),
		types.EntityHandler:          entities.Handler(),
		types.EntityNode:             entities.Node(),
		types.EntitySchedule:         entities.Schedule(),
		types.EntitySettings:         entities.Settings(),
		types.EntitySource:           entities.Source(),
		types.EntityTask:             entities.Task(),
		types.EntityUser:             entities.User(),
		types.EntityVirtualAssistant: entities.VirtualAssistant(),
		types.EntityVirtualDevice:    entities.VirtualDevice(),
		types.EntityServiceToken:     entities.ServiceToken(),
	}

	return funcMap, nil
}

func GetDirectories(includeSecureShare, includeInsecureShare bool) (map[string]string, error) {
	directories := map[string]string{}

	// include firmware directory
	firmwareDirSrc := types.GetEnvString(types.ENV_DIR_DATA_FIRMWARE)
	if firmwareDirSrc == "" {
		return nil, fmt.Errorf("environment '%s' not set", types.ENV_DIR_DATA_FIRMWARE)
	}
	directories[config.DirectoryDataFirmware] = firmwareDirSrc

	// include secure share
	if includeSecureShare {
		secureShareDirSrc := types.GetEnvString(types.ENV_DIR_SHARE_SECURE)
		if secureShareDirSrc == "" {
			return nil, fmt.Errorf("environment '%s' not set", types.ENV_DIR_SHARE_SECURE)
		}
		directories[config.DirectorySecureShare] = secureShareDirSrc
	}

	// include in-secure share
	if includeInsecureShare {
		inSecureShareDirSrc := types.GetEnvString(types.ENV_DIR_SHARE_INSECURE)
		if inSecureShareDirSrc == "" {
			return nil, fmt.Errorf("environment '%s' not set", types.ENV_DIR_SHARE_INSECURE)
		}
		directories[config.DirectoryInsecureShare] = inSecureShareDirSrc

	}

	return directories, nil
}

// verify MyControllerDataTransformationExportFunc, implements DataTransformerFunc
var _ backupPlugin.DataTransformerFunc = MyControllerDataTransformationExportFunc

func MyControllerDataTransformationExportFunc(logger *zap.Logger, entityName string, data interface{}, storageExportType string) (interface{}, error) {

	var modifiedData interface{}
	switch entityName {

	case types.EntityUser:
		// update "User" to "UserWithPassword" to keep the password in json export
		if users, ok := data.(*[]userTY.User); ok {
			usersWithPasswd := make([]userTY.UserWithPassword, len(*users))
			for _index, user := range *users {
				usersWithPasswd[_index] = userTY.UserWithPassword(user)
			}
			if len(usersWithPasswd) > 0 {
				modifiedData = usersWithPasswd
			}
		} else {
			return nil, fmt.Errorf("error on converting the data to user slice, continue with default data type. inputType:%s", fmt.Sprintf("%T", data))
		}

	}

	if modifiedData != nil {
		return modifiedData, nil
	}

	return data, nil
}
