package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	convertorUtils "github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	filterUtils "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	"github.com/nleeper/goment"
	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

// output types
const (
	OutputConsole     = "console"
	OutputConsoleWide = "wide" // same as console, prints additional headers
	OutputYAML        = "yaml"
	OutputJSON        = "json"
)

type Header struct {
	Title        string
	ValuePath    string
	DisplayStyle string
	IsWide       bool
}

const (
	DisplayStyleRelativeTime = "relative_time"
)

func Print(out io.Writer, headers []Header, data interface{}, hideHeader bool, output string, pretty bool) {
	switch output {
	case OutputConsole, OutputConsoleWide:
		dataConsole, ok := data.([]interface{})
		if !ok {
			fmt.Fprintln(out, "data not in table format")
			return
		}
		wideEnabled := false
		if output == OutputConsoleWide {
			wideEnabled = true
		}
		PrintConsole(out, headers, dataConsole, hideHeader, wideEnabled)
		return

	case OutputJSON:
		var jsonBytes []byte
		var err error
		if pretty {
			jsonBytes, err = json.MarshalIndent(data, "", " ")
		} else {
			jsonBytes, err = json.Marshal(data)
		}
		if err != nil {
			fmt.Println("error on converting to json", err)
			return
		}
		fmt.Fprint(out, string(jsonBytes))

	case OutputYAML:
		bytes, err := yaml.Marshal(data)
		if err != nil {
			fmt.Println("error on converting to yaml", err)
			return
		}
		fmt.Fprint(out, string(bytes))
	}
}

func PrintConsole(out io.Writer, headers []Header, data []interface{}, hideHeader, wideEnabled bool) {
	// update header
	updatedHeaders := []string{}
	for _, header := range headers {
		if !wideEnabled && header.IsWide {
			continue
		}
		updatedHeaders = append(updatedHeaders, header.Title)
	}

	// convert the data
	rows := make([][]string, 0)
	for index := range data {
		structData := data[index]
		row := make([]string, 0)
		for _, header := range headers {
			if !wideEnabled && header.IsWide {
				continue
			}

			valuePath := header.ValuePath
			if valuePath == "" {
				valuePath = header.Title
			}
			_, value, err := filterUtils.GetValueByKeyPath(structData, valuePath)
			if err != nil {
				errValue := err.Error()
				if strings.HasPrefix(err.Error(), "key not found") {
					errValue = ""
				}
				row = append(row, errValue)
				continue
			}

			var rowValue interface{}
			if value != nil {
				switch _value := value.(type) {
				case time.Time:
					if !_value.IsZero() {
						if header.DisplayStyle == DisplayStyleRelativeTime {
							g, err := goment.New(_value.UnixNano())
							if err != nil {
								rowValue = err.Error()
							} else {
								rowValue = g.FromNow()
							}
						}
					} else {
						rowValue = ""
					}

				case cmap.CustomStringMap:
					stringLabels := []string{}
					for k, v := range _value {
						stringLabels = append(stringLabels, fmt.Sprintf("%s=%s", k, v))
					}
					rowValue = strings.Join(stringLabels, ",")

				}

				if rowValue == nil {
					rowValue = fmt.Sprintf("%v", value)
				}
			}
			row = append(row, convertorUtils.ToString(rowValue))
		}
		rows = append(rows, row)
	}

	table := tablewriter.NewWriter(os.Stdout)
	if !hideHeader {
		table.SetHeader(updatedHeaders)
		table.SetAutoFormatHeaders(true)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	}
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("   ") // pad with space
	table.SetNoWhiteSpace(true)
	table.AppendBulk(rows) // Add Bulk Data
	table.Render()
}
