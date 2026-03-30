package helper

import (
	"testing"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	sfTY "github.com/mycontroller-org/server/v2/pkg/types/service_filter"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"github.com/stretchr/testify/assert"
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

func TestCompareTime(t *testing.T) {
	base := time.Date(2024, time.June, 15, 12, 0, 0, 0, time.UTC)
	earlier := time.Date(2024, time.June, 10, 12, 0, 0, 0, time.UTC)
	later := time.Date(2024, time.June, 20, 12, 0, 0, 0, time.UTC)

	baseStr := base.Format(time.RFC3339)
	earlierStr := earlier.Format(time.RFC3339)

	tests := []struct {
		name     string
		value    time.Time
		operator string
		filter   string
		expected bool
	}{
		{name: "eq match", value: base, operator: storageTY.OperatorEqual, filter: baseStr, expected: true},
		{name: "eq no match", value: base, operator: storageTY.OperatorEqual, filter: earlierStr, expected: false},
		{name: "ne match", value: base, operator: storageTY.OperatorNotEqual, filter: earlierStr, expected: true},
		{name: "ne no match", value: base, operator: storageTY.OperatorNotEqual, filter: baseStr, expected: false},
		{name: "gt match", value: later, operator: storageTY.OperatorGreaterThan, filter: baseStr, expected: true},
		{name: "gt equal no match", value: base, operator: storageTY.OperatorGreaterThan, filter: baseStr, expected: false},
		{name: "gt earlier no match", value: earlier, operator: storageTY.OperatorGreaterThan, filter: baseStr, expected: false},
		{name: "gte match equal", value: base, operator: storageTY.OperatorGreaterThanEqual, filter: baseStr, expected: true},
		{name: "gte match greater", value: later, operator: storageTY.OperatorGreaterThanEqual, filter: baseStr, expected: true},
		{name: "gte no match", value: earlier, operator: storageTY.OperatorGreaterThanEqual, filter: baseStr, expected: false},
		{name: "lt match", value: earlier, operator: storageTY.OperatorLessThan, filter: baseStr, expected: true},
		{name: "lt equal no match", value: base, operator: storageTY.OperatorLessThan, filter: baseStr, expected: false},
		{name: "lt later no match", value: later, operator: storageTY.OperatorLessThan, filter: baseStr, expected: false},
		{name: "lte match equal", value: base, operator: storageTY.OperatorLessThanEqual, filter: baseStr, expected: true},
		{name: "lte match less", value: earlier, operator: storageTY.OperatorLessThanEqual, filter: baseStr, expected: true},
		{name: "lte no match", value: later, operator: storageTY.OperatorLessThanEqual, filter: baseStr, expected: false},
		{name: "invalid filter value", value: base, operator: storageTY.OperatorEqual, filter: "not-a-time", expected: false},
		{name: "unknown operator", value: base, operator: "unknown", filter: baseStr, expected: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := CompareTime(tc.value, tc.operator, tc.filter)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsMatchingTimeField(t *testing.T) {
	type event struct {
		ID        string
		CreatedAt time.Time
	}

	base := time.Date(2024, time.June, 15, 12, 0, 0, 0, time.UTC)
	earlier := time.Date(2024, time.June, 10, 12, 0, 0, 0, time.UTC)
	later := time.Date(2024, time.June, 20, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		entity   interface{}
		filters  []storageTY.Filter
		expected bool
	}{
		{
			name:   "gt match",
			entity: &event{ID: "1", CreatedAt: later},
			filters: []storageTY.Filter{
				{Key: "createdAt", Operator: storageTY.OperatorGreaterThan, Value: base.Format(time.RFC3339)},
			},
			expected: true,
		},
		{
			name:   "gt no match",
			entity: &event{ID: "2", CreatedAt: earlier},
			filters: []storageTY.Filter{
				{Key: "createdAt", Operator: storageTY.OperatorGreaterThan, Value: base.Format(time.RFC3339)},
			},
			expected: false,
		},
		{
			name:   "lt match",
			entity: &event{ID: "3", CreatedAt: earlier},
			filters: []storageTY.Filter{
				{Key: "createdAt", Operator: storageTY.OperatorLessThan, Value: base.Format(time.RFC3339)},
			},
			expected: true,
		},
		{
			name:   "lte match equal",
			entity: &event{ID: "4", CreatedAt: base},
			filters: []storageTY.Filter{
				{Key: "createdAt", Operator: storageTY.OperatorLessThanEqual, Value: base.Format(time.RFC3339)},
			},
			expected: true,
		},
		{
			name:   "gte no match",
			entity: &event{ID: "5", CreatedAt: earlier},
			filters: []storageTY.Filter{
				{Key: "createdAt", Operator: storageTY.OperatorGreaterThanEqual, Value: base.Format(time.RFC3339)},
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsMatching(tc.entity, tc.filters)
			assert.Equal(t, tc.expected, result)
		})
	}
}
