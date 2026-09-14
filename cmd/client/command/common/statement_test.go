package common

import (
	"testing"

	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOptionalStatement(t *testing.T) {
	_, provided, err := ParseOptionalStatement("", nil, nil)
	require.NoError(t, err)
	assert.False(t, provided)

	_, _, err = ParseOptionalStatement("", []string{"get"}, nil)
	require.Error(t, err)

	_, _, err = ParseOptionalStatement("", nil, []string{"node:*"})
	require.Error(t, err)

	sts, provided, err := ParseOptionalStatement("deny", []string{"*"}, []string{"settings"})
	require.NoError(t, err)
	require.True(t, provided)
	require.Len(t, sts, 1)
	assert.Equal(t, policyTY.EffectDeny, sts[0].Effect)
	assert.Equal(t, []string{"*"}, sts[0].Actions)
	assert.Equal(t, []string{"settings"}, sts[0].Resources)
}
