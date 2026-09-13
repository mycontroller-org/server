package get

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleSettings() map[string]interface{} {
	return map[string]interface{}{
		"language": "en",
		"geoLocation": map[string]interface{}{
			"autoUpdate":   true,
			"locationName": "Berlin",
			"latitude":     52.5,
		},
		"login": map[string]interface{}{
			"message": "hello",
		},
	}
}

func TestLookupSettingsPathRoot(t *testing.T) {
	spec := sampleSettings()
	got, err := lookupSettingsPath(spec, "")
	require.NoError(t, err)
	nested, ok := got.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, []string{"geoLocation", "language", "login"}, sortedKeys(nested))
}

func TestLookupSettingsPathMap(t *testing.T) {
	got, err := lookupSettingsPath(sampleSettings(), "geoLocation")
	require.NoError(t, err)
	nested := got.(map[string]interface{})
	assert.Equal(t, true, nested["autoUpdate"])
	assert.Equal(t, 52.5, nested["latitude"])
}

func TestLookupSettingsPathLeaf(t *testing.T) {
	got, err := lookupSettingsPath(sampleSettings(), "geoLocation.latitude")
	require.NoError(t, err)
	assert.Equal(t, 52.5, got)
}

func TestLookupSettingsPathMissing(t *testing.T) {
	_, err := lookupSettingsPath(sampleSettings(), "geoLocation.missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not present")
}

func TestFlattenSettingsUnderPrefix(t *testing.T) {
	rows := flattenSettings(sampleSettings()["geoLocation"].(map[string]interface{}), "geoLocation")
	require.Len(t, rows, 3)
	assert.Equal(t, "geoLocation.autoUpdate", rows[0].(settingsRow).Key)
	assert.Equal(t, "true", rows[0].(settingsRow).Value)
	assert.Equal(t, "geoLocation.latitude", rows[1].(settingsRow).Key)
	assert.Equal(t, "52.5", rows[1].(settingsRow).Value)
}
