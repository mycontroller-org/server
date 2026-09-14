package alias

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateAliasName(t *testing.T) {
	require.NoError(t, validateAliasName("home"))
	require.NoError(t, validateAliasName("prod-1"))
	require.NoError(t, validateAliasName("Lab_A"))
	require.Error(t, validateAliasName(""))
	require.Error(t, validateAliasName("1prod"))
	require.Error(t, validateAliasName("has space"))
	require.Error(t, validateAliasName("get"))
	require.Error(t, validateAliasName("GET"))
	assert.Contains(t, validateAliasName("apply").Error(), "reserved")
	assert.Contains(t, validateAliasName("add").Error(), "reserved")
	assert.Contains(t, validateAliasName("create").Error(), "reserved")
	assert.Contains(t, validateAliasName("update").Error(), "reserved")
}
