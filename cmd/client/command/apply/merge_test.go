package apply

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeepMergeMapsKeepsMissingKeys(t *testing.T) {
	base := map[string]interface{}{
		"name": "old",
		"labels": map[string]interface{}{
			"keep": "yes",
			"room": "kitchen",
		},
		"others": map[string]interface{}{
			"note": "stay",
			"nested": map[string]interface{}{
				"a": "1",
				"b": "2",
			},
		},
	}
	overlay := map[string]interface{}{
		"name": "new",
		"labels": map[string]interface{}{
			"room": "living",
		},
		"others": map[string]interface{}{
			"nested": map[string]interface{}{
				"b": "9",
			},
		},
	}
	merged, ok := deepMergeValue(base, overlay).(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "new", merged["name"])
	labels := merged["labels"].(map[string]interface{})
	assert.Equal(t, "yes", labels["keep"])
	assert.Equal(t, "living", labels["room"])
	others := merged["others"].(map[string]interface{})
	assert.Equal(t, "stay", others["note"])
	nested := others["nested"].(map[string]interface{})
	assert.Equal(t, "1", nested["a"])
	assert.Equal(t, "9", nested["b"])
}

func TestDeepMergeArrayByID(t *testing.T) {
	base := []interface{}{
		map[string]interface{}{"id": "a", "name": "old", "keep": "yes"},
		map[string]interface{}{"id": "b", "name": "b"},
	}
	overlay := []interface{}{
		map[string]interface{}{"id": "a", "name": "new"},
		map[string]interface{}{"id": "c", "name": "c"},
	}
	merged, ok := deepMergeValue(base, overlay).([]interface{})
	require.True(t, ok)
	require.Len(t, merged, 3)
	first := merged[0].(map[string]interface{})
	assert.Equal(t, "new", first["name"])
	assert.Equal(t, "yes", first["keep"])
	assert.Equal(t, "b", merged[1].(map[string]interface{})["id"])
	assert.Equal(t, "c", merged[2].(map[string]interface{})["id"])
}

func TestDeepMergeArrayByField(t *testing.T) {
	base := []interface{}{
		map[string]interface{}{"field": "temp", "unit": "C", "name": "old"},
		map[string]interface{}{"field": "hum", "unit": "%"},
	}
	overlay := []interface{}{
		map[string]interface{}{"field": "temp", "name": "Temperature"},
	}
	merged, ok := deepMergeValue(base, overlay).([]interface{})
	require.True(t, ok)
	require.Len(t, merged, 2)
	temp := merged[0].(map[string]interface{})
	assert.Equal(t, "Temperature", temp["name"])
	assert.Equal(t, "C", temp["unit"])
	assert.Equal(t, "hum", merged[1].(map[string]interface{})["field"])
}

func TestDeepMergeScalarArrayReplaces(t *testing.T) {
	merged := deepMergeValue([]interface{}{"a", "b"}, []interface{}{"c"})
	assert.Equal(t, []interface{}{"c"}, merged)
}
