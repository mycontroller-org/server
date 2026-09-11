package api

import (
	"testing"

	dataRepoTY "github.com/mycontroller-org/server/v2/pkg/types/data_repository"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyJSONPathNestedScript(t *testing.T) {
	in := &fieldTY.Field{
		ID:        "id-1",
		GatewayID: "gw",
		NodeID:    "n",
		SourceID:  "s",
		FieldID:   "temp",
		Name:      "Temperature",
	}
	out := &fieldTY.Field{}
	err := applyJSONPath(in, "formatter.onReceive", "return value;", true, out)
	require.NoError(t, err)
	assert.Equal(t, "id-1", out.ID)
	assert.Equal(t, "Temperature", out.Name)
	assert.Equal(t, "return value;", out.Formatter.OnReceive)
}

func TestApplyJSONPathDataRepository(t *testing.T) {
	in := &dataRepoTY.Config{
		ID:   "ota",
		Data: map[string]interface{}{"disabled": false},
	}
	out := &dataRepoTY.Config{}
	err := applyJSONPath(in, "data.onConfig", "var x = 1;", true, out)
	require.NoError(t, err)
	assert.Equal(t, "ota", out.ID)
	assert.Equal(t, "var x = 1;", out.Data["onConfig"])
	assert.Equal(t, false, out.Data["disabled"])
}

func TestApplyJSONPathDecodesJSONValues(t *testing.T) {
	type sample struct {
		Enabled bool    `json:"enabled"`
		Count   float64 `json:"count"`
		Name    string  `json:"name"`
	}
	in := &sample{Name: "keep"}
	out := &sample{}
	require.NoError(t, applyJSONPath(in, "enabled", "true", false, out))
	assert.True(t, out.Enabled)
	assert.Equal(t, "keep", out.Name)

	out = &sample{}
	require.NoError(t, applyJSONPath(in, "count", "3", false, out))
	assert.Equal(t, float64(3), out.Count)

	out = &sample{}
	require.NoError(t, applyJSONPath(in, "name", "true", true, out))
	assert.Equal(t, "true", out.Name)
}
