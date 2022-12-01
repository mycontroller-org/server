package get

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	rootCmd "github.com/mycontroller-org/server/v2/cmd/client/command/root"
	"github.com/mycontroller-org/server/v2/cmd/client/printer"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"

	"github.com/spf13/cobra"
)

var (
	limit        uint64
	orderKey     string
	orderBy      string
	filterString string
)

type ListFunc func(queryParams map[string]interface{}) (*storageTY.Result, error)

func init() {
	rootCmd.Cmd.AddCommand(getCmd)
	getCmd.PersistentFlags().Uint64Var(&limit, "limit", 10, "Limits number of rows")
	getCmd.PersistentFlags().StringVar(&orderKey, "order-key", "id", "order the result by this key")
	getCmd.PersistentFlags().StringVar(&orderBy, "order-by", storageTY.SortByASC, "order the result. options: desc, asc")
	getCmd.PersistentFlags().StringVar(&filterString, "filter", "", "filter the result. comma separated key=value")
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Prints the requested resources",
	PreRun: func(cmd *cobra.Command, args []string) {
		rootCmd.UpdateStreams(cmd)
	},
}

func getFilters() []storageTY.Filter {
	if filterString == "" {
		return []storageTY.Filter{}
	}

	filters := []storageTY.Filter{}
	filtersRaw := strings.Split(filterString, ",")
	for _, filterMap := range filtersRaw {
		filterArray := strings.SplitN(filterMap, "=", 2)
		if len(filterArray) != 2 {
			continue
		}
		filters = append(filters, storageTY.Filter{
			Key:      strings.TrimSpace(filterArray[0]),
			Value:    strings.TrimSpace(filterArray[1]),
			Operator: storageTY.OperatorRegex,
		})
	}

	return filters
}

func GetQueryParams() (map[string]interface{}, error) {
	limit := limit
	pageOffset := uint64(0)
	sortBy := []storageTY.Sort{{OrderBy: orderBy, Field: orderKey}}

	sortByBytes, err := json.Marshal(sortBy)
	if err != nil {
		return nil, err
	}
	filtersBytes, err := json.Marshal(getFilters())
	if err != nil {
		return nil, err
	}
	queryParams := map[string]interface{}{
		"limit":  limit,
		"offset": pageOffset,
		"filter": string(filtersBytes),
		"sortBy": string(sortByBytes),
	}

	return queryParams, nil
}

func executeGetCmd(headers []printer.Header, listFunc ListFunc, dataType interface{}) {
	queryParams, err := GetQueryParams()
	if err != nil {
		fmt.Fprintf(rootCmd.IOStreams.ErrOut, "error:%s\n", err)
		return
	}

	result, err := listFunc(queryParams)
	if err != nil {
		fmt.Fprintf(rootCmd.IOStreams.ErrOut, "error:%s\n", err)
		return
	}
	res, ok := result.Data.([]interface{})
	if !ok {
		fmt.Fprintf(rootCmd.IOStreams.ErrOut, "invalid response type:%T\n", result.Data)
		return
	}

	if len(res) == 0 {
		fmt.Fprintln(rootCmd.IOStreams.Out, "No resource found")
		return
	}

	rows := make([]interface{}, 0)
	for _, rawData := range res {
		data, ok := rawData.(map[string]interface{})
		if !ok {
			continue
		}
		item := reflect.New(reflect.TypeOf(dataType)).Interface()
		err = utils.MapToStruct(utils.TagNameJSON, data, item)
		if err != nil {
			fmt.Fprintf(rootCmd.IOStreams.ErrOut, "error on map to struct. %s", err.Error())
			continue
		}

		rows = append(rows, item)
	}

	printer.Print(rootCmd.IOStreams.Out, headers, rows, rootCmd.HideHeader, rootCmd.OutputFormat, rootCmd.Pretty)
}
