package action

import (
	"fmt"

	dataRepoAPI "github.com/mycontroller-org/backend/v2/pkg/api/data_repository"
	"github.com/mycontroller-org/backend/v2/pkg/json"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	"github.com/tidwall/sjson"
)

func toDataRepository(id, selector, value string) error {
	dataRepo, err := dataRepoAPI.GetByID(id)
	if err != nil {
		return err
	}
	if dataRepo.ReadOnly {
		return fmt.Errorf("'%s' is a readonly data repository", id)
	}

	dataBytes, err := json.Marshal(dataRepo.Data)
	if err != nil {
		return err
	}

	jsonString, err := sjson.Set(string(dataBytes), selector, value)
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
