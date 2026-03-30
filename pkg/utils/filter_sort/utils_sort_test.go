package helper

import (
	"testing"
	"time"

	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"github.com/stretchr/testify/assert"
)

func TestNaturalStringLess(t *testing.T) {
	testCases := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{
			name:     "simple numeric comparison",
			a:        "file2",
			b:        "file10",
			expected: true,
		},
		{
			name:     "simple numeric comparison reversed",
			a:        "file10",
			b:        "file2",
			expected: false,
		},
		{
			name:     "equal strings",
			a:        "file1",
			b:        "file1",
			expected: false,
		},
		{
			name:     "no numbers - alphabetical",
			a:        "apple",
			b:        "banana",
			expected: true,
		},
		{
			name:     "no numbers - alphabetical reversed",
			a:        "banana",
			b:        "apple",
			expected: false,
		},
		{
			name:     "multiple digits",
			a:        "item100",
			b:        "item99",
			expected: false,
		},
		{
			name:     "multiple numbers in string",
			a:        "version2.10.5",
			b:        "version2.9.5",
			expected: false,
		},
		{
			name:     "prefix with numbers",
			a:        "10file",
			b:        "2file",
			expected: false,
		},
		{
			name:     "mixed alphanumeric",
			a:        "test1abc2",
			b:        "test1abc10",
			expected: true,
		},
		{
			name:     "one with number one without",
			a:        "file",
			b:        "file2",
			expected: true,
		},
		{
			name:     "different lengths same prefix",
			a:        "test",
			b:        "test123",
			expected: true,
		},
		{
			name:     "zero padding",
			a:        "file001",
			b:        "file2",
			expected: true,
		},
		{
			name:     "large numbers",
			a:        "item999",
			b:        "item1000",
			expected: true,
		},
		{
			name:     "unicode characters",
			a:        "文件2",
			b:        "文件10",
			expected: true,
		},
		{
			name:     "spaces in string",
			a:        "file 2",
			b:        "file 10",
			expected: true,
		},
		{
			name:     "dash separator with numbers",
			a:        "file-2",
			b:        "file-10",
			expected: true, // compares 'file-' prefix, then 2 < 10 numerically
		},
		{
			name:     "empty strings",
			a:        "",
			b:        "file",
			expected: true,
		},
		{
			name:     "both empty strings",
			a:        "",
			b:        "",
			expected: false,
		},
		{
			name:     "only numbers",
			a:        "123",
			b:        "45",
			expected: false,
		},
		{
			name:     "special characters",
			a:        "file#2",
			b:        "file#10",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := naturalStringLess(tc.a, tc.b)
			assert.Equal(t, tc.expected, result, "naturalStringLess(%q, %q) should return %v", tc.a, tc.b, tc.expected)
		})
	}
}

func TestGetSortByKeyPath_TimeSort(t *testing.T) {
	older := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	newer := time.Date(2024, time.February, 2, 3, 4, 5, 0, time.UTC)

	type sortTestEntity struct {
		ID           string
		StartDate    time.Time
		DeliveryDate time.Time
		Name         string
	}

	testCases := []struct {
		name     string
		keyPath  string
		orderBy  string
		data     []interface{}
		expected []string
	}{
		{
			name:    "ascending lower case",
			keyPath: "startDate",
			orderBy: storageTY.SortByASC,
			data: []interface{}{
				&sortTestEntity{ID: "2", StartDate: newer},
				&sortTestEntity{ID: "1", StartDate: older},
			},
			expected: []string{"1", "2"},
		},
		{
			name:    "descending lower case",
			keyPath: "deliveryDate",
			orderBy: storageTY.SortByDESC,
			data: []interface{}{
				&sortTestEntity{ID: "1", DeliveryDate: older},
				&sortTestEntity{ID: "2", DeliveryDate: newer},
			},
			expected: []string{"2", "1"},
		},
		{
			name:    "ascending upper case",
			keyPath: "startDate",
			orderBy: "ASC",
			data: []interface{}{
				&sortTestEntity{ID: "2", StartDate: newer},
				&sortTestEntity{ID: "1", StartDate: older},
			},
			expected: []string{"1", "2"},
		},
		{
			name:    "descending upper case",
			keyPath: "deliveryDate",
			orderBy: "DESC",
			data: []interface{}{
				&sortTestEntity{ID: "1", DeliveryDate: older},
				&sortTestEntity{ID: "2", DeliveryDate: newer},
			},
			expected: []string{"2", "1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sorted := GetSortByKeyPath(tc.keyPath, tc.orderBy, tc.data)
			actual := []string{
				sorted[0].(*sortTestEntity).ID,
				sorted[1].(*sortTestEntity).ID,
			}
			assert.Equal(t, tc.expected, actual)
		})
	}
}
