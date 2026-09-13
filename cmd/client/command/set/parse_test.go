package set

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSetArgsFile(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "script.js")
	require.NoError(t, os.WriteFile(name, []byte("return 1;"), 0o600))

	selectors, keyPath, value, err := parseSetArgs([]string{"mysensor.1.dht.temp", "formatter.onReceive"}, name, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"mysensor.1.dht.temp"}, selectors)
	assert.Equal(t, "formatter.onReceive", keyPath)
	assert.Equal(t, "return 1;", value)
}

func TestParseSetArgsPathFlagAndFile(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "onConfig.js")
	require.NoError(t, os.WriteFile(name, []byte("var x = 1;"), 0o600))

	selectors, keyPath, value, err := parseSetArgs([]string{"ota_stm32_ab"}, name, "data.onConfig")
	require.NoError(t, err)
	assert.Equal(t, []string{"ota_stm32_ab"}, selectors)
	assert.Equal(t, "data.onConfig", keyPath)
	assert.Equal(t, "var x = 1;", value)
}

func TestParseSetArgsInlinePath(t *testing.T) {
	selectors, keyPath, value, err := parseSetArgs([]string{"gw1", "description", "USB"}, "", "")
	require.NoError(t, err)
	assert.Equal(t, []string{"gw1"}, selectors)
	assert.Equal(t, "description", keyPath)
	assert.Equal(t, "USB", value)
}

func TestParseSetArgsFieldNestedInline(t *testing.T) {
	selectors, keyPath, value, err := parseSetArgs([]string{"mysensor.1.dht.temp", "formatter.onReceive", "return v;"}, "", "")
	require.NoError(t, err)
	assert.Equal(t, []string{"mysensor.1.dht.temp"}, selectors)
	assert.Equal(t, "formatter.onReceive", keyPath)
	assert.Equal(t, "return v;", value)
}

func TestParseSetArgsMissing(t *testing.T) {
	_, _, _, err := parseSetArgs([]string{"gw1"}, "", "")
	require.Error(t, err)
}

func TestParseSetArgsKeyPathWithoutValue(t *testing.T) {
	_, _, _, err := parseSetArgs([]string{"gw1.1.1.V_CUSTOM", "formatter.onReceive"}, "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "value or --file is required")
	assert.Contains(t, err.Error(), "formatter.onReceive")
}

func TestParseSettingsArgs(t *testing.T) {
	path, value, err := parseSettingsArgs([]string{"language", "en"}, "", "")
	require.NoError(t, err)
	assert.Equal(t, "language", path)
	assert.Equal(t, "en", value)

	_, _, err = parseSettingsArgs([]string{"geoLocation.autoUpdate"}, "", "")
	require.Error(t, err)

	path, value, err = parseSettingsArgs([]string{"true"}, "", "geoLocation.autoUpdate")
	require.NoError(t, err)
	assert.Equal(t, "geoLocation.autoUpdate", path)
	assert.Equal(t, "true", value)
}

func TestParseSetArgsBareNameRequiresValue(t *testing.T) {
	_, _, _, err := parseSetArgs([]string{"gw1.1.1.V_CUSTOM", "name"}, "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "resource id, key path, and value are required")
}
