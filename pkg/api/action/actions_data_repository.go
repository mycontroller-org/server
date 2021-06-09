package action

import (
	"strings"

	dataRepoAPI "github.com/mycontroller-org/backend/v2/pkg/api/data_repository"
	"github.com/mycontroller-org/backend/v2/pkg/json"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	"github.com/tidwall/sjson"
	"go.uber.org/zap"
)

func toDataRepository(id, keyPath, value string) error {
	dataRepo, err := dataRepoAPI.GetByID(id)
	if err != nil {
		return err
	}
	if dataRepo.ReadOnly {
		zap.L().Info("update failed: trying update a readonly repository", zap.String("id", id), zap.String("keyPath", keyPath), zap.String("value", value))
		return nil
	}

	dataBytes, err := json.Marshal(dataRepo.Data)
	if err != nil {
		return err
	}

	var finalValue interface{}

	// if the supplied string is a json, convert it to map of interface and add it in to the data repository
	// convert the value as map interface
	var valueObject interface{}
	if strings.HasPrefix(value, "[") { // it is an array of objects
		valueObject = make([]interface{}, 0)
	} else {
		valueObject = make(map[string]interface{})
	}

	updateValue := value
	// ugly hack to remove the escaped double quote in json, conversion in template
	if strings.Contains(value, "&#34;") {
		updateValue = strings.ReplaceAll(value, "&#34;", "\"")
	}
	err = json.Unmarshal([]byte(updateValue), &valueObject)
	if err != nil {
		finalValue = value
	} else {
		finalValue = valueObject
	}

	// inject the final value
	jsonString, err := sjson.Set(string(dataBytes), keyPath, finalValue)
	if err != nil {
		return err
	}

	var finalData cmap.CustomMap
	finalData = finalData.Init()
	err = json.Unmarshal([]byte(jsonString), &finalData)
	if err != nil {
		return err
	}

	dataRepo.Data = finalData
	return dataRepoAPI.Save(dataRepo)
}
