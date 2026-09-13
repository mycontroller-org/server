package get

import (
	"testing"

	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	"github.com/stretchr/testify/assert"
)

func TestFormatStatements(t *testing.T) {
	assert.Equal(t, "-", formatStatements(nil))
	assert.Equal(t, "-", formatStatements([]policyTY.Statement{}))
	assert.Equal(t, "Allow get,list node:*", formatStatements([]policyTY.Statement{{
		Effect:    policyTY.EffectAllow,
		Actions:   []string{"get", "list"},
		Resources: []string{"node:*"},
	}}))
	assert.Equal(t, "Allow get node:*; Deny * settings", formatStatements([]policyTY.Statement{
		{Effect: policyTY.EffectAllow, Actions: []string{"get"}, Resources: []string{"node:*"}},
		{Effect: policyTY.EffectDeny, Actions: []string{"*"}, Resources: []string{"settings"}},
	}))
}
