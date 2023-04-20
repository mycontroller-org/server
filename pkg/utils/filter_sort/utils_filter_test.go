package helper

import (
	"testing"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	sfTY "github.com/mycontroller-org/server/v2/pkg/types/service_filter"
	"github.com/stretchr/testify/require"
)

func TestIsMine(t *testing.T) {
	testData := []struct {
		name           string
		filter         sfTY.ServiceFilter
		targetType     string
		targetID       string
		targetLabels   cmap.CustomStringMap
		expectedResult bool
	}{
		{
			name:           "empty match",
			filter:         sfTY.ServiceFilter{},
			expectedResult: true,
		},
		{
			name:       "targetType match",
			targetType: "type-123",
			filter: sfTY.ServiceFilter{
				MatchAll: false,
				Types:    []string{"type-123"},
			},
			expectedResult: true,
		},
		{
			name:     "targetID match",
			targetID: "type-123",
			filter: sfTY.ServiceFilter{
				MatchAll: false,
				IDs:      []string{"type-1", "type-123"},
			},
			expectedResult: true,
		},
		{
			name: "label match",
			targetLabels: cmap.CustomStringMap{
				"location": "india",
			},
			filter: sfTY.ServiceFilter{
				MatchAll: false,
				Labels: cmap.CustomStringMap{
					"location": "india",
				},
			},
			expectedResult: true,
		},
		{
			name:       "label or type match",
			targetType: "device1",
			targetLabels: cmap.CustomStringMap{
				"location": "india",
			},
			filter: sfTY.ServiceFilter{
				MatchAll: false,
				Types:    []string{"device2"},
				Labels: cmap.CustomStringMap{
					"location": "india",
				},
			},
			expectedResult: true,
		},
		{
			name: "type or label",
			targetLabels: cmap.CustomStringMap{
				"location": "india",
			},
			targetType: "mqtt",
			filter: sfTY.ServiceFilter{
				MatchAll: false,
				Types:    []string{"mqtt"},
				Labels: cmap.CustomStringMap{
					"location": "tn",
				},
			},
			expectedResult: true,
		},
		{
			name: "id or type or label",
			targetLabels: cmap.CustomStringMap{
				"location": "india",
			},
			targetType: "mqtt",
			targetID:   "hello123",
			filter: sfTY.ServiceFilter{
				MatchAll: false,
				Types:    []string{"serial"},
				IDs:      []string{"hello123"},
				Labels: cmap.CustomStringMap{
					"location": "tn",
				},
			},
			expectedResult: true,
		},
		{
			name: "match all",
			targetLabels: cmap.CustomStringMap{
				"location": "india",
			},
			targetType: "mqtt",
			targetID:   "hello123",
			filter: sfTY.ServiceFilter{
				MatchAll: true,
				Types:    []string{"mqtt"},
				IDs:      []string{"hello123"},
				Labels: cmap.CustomStringMap{
					"location": "india",
				},
			},
			expectedResult: true,
		},
		{
			name: "match all not match",
			targetLabels: cmap.CustomStringMap{
				"location": "india",
			},
			targetType: "mqtt",
			targetID:   "hello123",
			filter: sfTY.ServiceFilter{
				MatchAll: true,
				Types:    []string{"serial"},
				IDs:      []string{"hello123"},
				Labels: cmap.CustomStringMap{
					"location": "india",
				},
			},
			expectedResult: false,
		},
		{
			name: "match all no match all",
			targetLabels: cmap.CustomStringMap{
				"location": "india",
			},
			targetType: "mqtt",
			targetID:   "hello123",
			filter: sfTY.ServiceFilter{
				MatchAll: true,
				Types:    []string{"serial"},
				IDs:      []string{"hi123"},
				Labels: cmap.CustomStringMap{
					"location": "tn",
				},
			},
			expectedResult: false,
		},
		{
			name:       "target type not match",
			targetType: "my-id-123",
			filter: sfTY.ServiceFilter{
				MatchAll: false,
				Types:    []string{"type-123"},
			},
			expectedResult: false,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			actualResult := IsMine(&test.filter, test.targetType, test.targetID, test.targetLabels)
			require.Equal(t, test.expectedResult, actualResult)
		})
	}

}
