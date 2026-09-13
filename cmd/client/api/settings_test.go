package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeSettingMaps(t *testing.T) {
	base := map[string]interface{}{
		"language": "en",
		"geoLocation": map[string]interface{}{
			"autoUpdate":   true,
			"locationName": "here",
		},
	}
	overlay := map[string]interface{}{
		"language": "de",
		"geoLocation": map[string]interface{}{
			"latitude": 1.5,
		},
	}
	merged := mergeSettingMaps(base, overlay)
	assert.Equal(t, "de", merged["language"])
	geo := merged["geoLocation"].(map[string]interface{})
	assert.Equal(t, true, geo["autoUpdate"])
	assert.Equal(t, "here", geo["locationName"])
	assert.Equal(t, 1.5, geo["latitude"])
}

func TestApplyJSONPathOnSettingsSpec(t *testing.T) {
	spec := map[string]interface{}{
		"language": "en",
		"geoLocation": map[string]interface{}{
			"autoUpdate": false,
		},
	}
	out := map[string]interface{}{}
	require.NoError(t, applyJSONPath(spec, "geoLocation.autoUpdate", "true", false, &out))
	geo := out["geoLocation"].(map[string]interface{})
	assert.Equal(t, true, geo["autoUpdate"])
	assert.Equal(t, "en", out["language"])
}
