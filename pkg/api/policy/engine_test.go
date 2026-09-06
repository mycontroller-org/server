package policy

import (
	"testing"

	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
)

func TestStatementsAllow(t *testing.T) {
	sts := []policyTY.Statement{
		{
			Effect:    policyTY.EffectAllow,
			Actions:   []string{policyTY.ActionGet, policyTY.ActionList},
			Resources: []string{"field:plant-room.sensor-01.*"},
		},
	}
	if !statementsAllow(sts, "get", "field:plant-room.sensor-01.climate.temp") {
		t.Fatal("expected allow get on field")
	}
	if statementsAllow(sts, "update", "field:plant-room.sensor-01.climate.temp") {
		t.Fatal("expected deny update")
	}
	if statementsAllow(sts, "get", "field:plant-room.other.climate.temp") {
		t.Fatal("expected deny other node")
	}
}

func TestStatementsDenyWins(t *testing.T) {
	sts := []policyTY.Statement{
		{
			Effect:    policyTY.EffectAllow,
			Actions:   []string{"*"},
			Resources: []string{"*"},
		},
		{
			Effect:    policyTY.EffectDeny,
			Actions:   []string{"*"},
			Resources: []string{"settings"},
		},
	}
	if !statementsAllow(sts, "get", "gateway:plant-room") {
		t.Fatal("expected allow gateway under *")
	}
	if statementsAllow(sts, "get", "settings") {
		t.Fatal("expected deny settings")
	}
	if statementsAllow(sts, "update", "settings") {
		t.Fatal("expected deny settings update")
	}
}

func TestStatementsDenyNamedResource(t *testing.T) {
	sts := []policyTY.Statement{
		{
			Effect:    policyTY.EffectAllow,
			Actions:   []string{policyTY.ActionGet, policyTY.ActionList},
			Resources: []string{"gateway:*"},
		},
		{
			Effect:    policyTY.EffectDeny,
			Actions:   []string{"*"},
			Resources: []string{"gateway:secret-gw"},
		},
	}
	if !statementsAllow(sts, "get", "gateway:plant-room") {
		t.Fatal("expected allow plant-room")
	}
	if statementsAllow(sts, "get", "gateway:secret-gw") {
		t.Fatal("expected deny secret-gw")
	}
}

// Deny node:mysensor.1 must cascade to fields/sources/metrics under that node.
func TestStatementsDenyNodeCascadesToChildren(t *testing.T) {
	sts := []policyTY.Statement{
		{
			Effect:  policyTY.EffectAllow,
			Actions: []string{"*"},
			Resources: []string{
				"node:mysensor.*",
				"source:mysensor.*",
				"field:mysensor.*",
				"metric:mysensor.*",
			},
		},
		{
			Effect:    policyTY.EffectDeny,
			Actions:   []string{"*"},
			Resources: []string{"node:mysensor.1"},
		},
	}
	if !statementsAllow(sts, "get", "node:mysensor.2") {
		t.Fatal("expected allow other node")
	}
	if statementsAllow(sts, "get", "node:mysensor.1") {
		t.Fatal("expected deny node itself")
	}
	if statementsAllow(sts, "get", "field:mysensor.1.s1.temp") {
		t.Fatal("expected deny field under denied node")
	}
	if statementsAllow(sts, "list", "field:mysensor.1.s1.temp") {
		t.Fatal("expected deny list field under denied node")
	}
	if statementsAllow(sts, "get", "source:mysensor.1.s1") {
		t.Fatal("expected deny source under denied node")
	}
	if statementsAllow(sts, "get", "metric:mysensor.1.s1.temp") {
		t.Fatal("expected deny metric under denied node")
	}
	if !statementsAllow(sts, "get", "field:mysensor.2.s1.temp") {
		t.Fatal("expected allow field under other node")
	}
}

func TestStatementsDenyGatewayCascades(t *testing.T) {
	sts := []policyTY.Statement{
		{
			Effect:    policyTY.EffectAllow,
			Actions:   []string{"*"},
			Resources: []string{"*"},
		},
		{
			Effect:    policyTY.EffectDeny,
			Actions:   []string{"*"},
			Resources: []string{"gateway:secret"},
		},
	}
	if statementsAllow(sts, "get", "node:secret.1") {
		t.Fatal("expected deny node under denied gateway")
	}
	if statementsAllow(sts, "get", "field:secret.1.s.f") {
		t.Fatal("expected deny field under denied gateway")
	}
	if !statementsAllow(sts, "get", "field:other.1.s.f") {
		t.Fatal("expected allow other gateway fields")
	}
}

// Allow cascades: Allow node:x grants field/source under x (device tree).
func TestStatementsAllowCascadesToChildren(t *testing.T) {
	sts := []policyTY.Statement{
		{
			Effect:    policyTY.EffectAllow,
			Actions:   []string{"*"},
			Resources: []string{"node:mysensor.1"},
		},
	}
	if !statementsAllow(sts, "get", "node:mysensor.1") {
		t.Fatal("expected allow node")
	}
	if !statementsAllow(sts, "get", "field:mysensor.1.s.f") {
		t.Fatal("allow on node must cascade to fields under that node")
	}
	if !statementsAllow(sts, "get", "source:mysensor.1.s") {
		t.Fatal("allow on node must cascade to sources under that node")
	}
	if statementsAllow(sts, "get", "field:mysensor.2.s.f") {
		t.Fatal("must not allow fields under other nodes")
	}
}

// Real user policy shape: allow gateway + node tree, deny one node.
func TestUserPolicyGatewayNodeOnlyWithDeny(t *testing.T) {
	sts := []policyTY.Statement{
		{
			Effect:  policyTY.EffectAllow,
			Actions: []string{"get", "list"},
			Resources: []string{
				"gateway:mysensor",
				"node:mysensor.*",
			},
		},
		{
			Effect:    policyTY.EffectDeny,
			Actions:   []string{"get", "list"},
			Resources: []string{"node:mysensor.1"},
		},
	}
	// siblings allowed via Allow cascade
	if !statementsAllow(sts, "get", "field:mysensor.2.s.temp") {
		t.Fatal("expected allow field under node 2")
	}
	if !statementsAllow(sts, "get", "source:mysensor.2.s") {
		t.Fatal("expected allow source under node 2")
	}
	if !statementsAllow(sts, "get", "node:mysensor.2") {
		t.Fatal("expected allow node 2")
	}
	// denied node and children blocked
	if statementsAllow(sts, "get", "node:mysensor.1") {
		t.Fatal("expected deny node 1")
	}
	if statementsAllow(sts, "get", "field:mysensor.1.s.temp") {
		t.Fatal("expected deny field under node 1")
	}
	if statementsAllow(sts, "get", "source:mysensor.1.s") {
		t.Fatal("expected deny source under node 1")
	}
	// other gateway not allowed
	if statementsAllow(sts, "get", "field:other.1.s.t") {
		t.Fatal("must not allow other gateway")
	}
}

func TestRestrictionsAllow(t *testing.T) {
	// empty = unrestricted (same as user ceiling)
	if !restrictionsAllow(nil, nil, "delete", "gateway:x") {
		t.Fatal("empty restrictions should allow")
	}
	if !restrictionsAllow([]string{"get", "list"}, []string{"field:gw.n.*"}, "get", "field:gw.n.s.f") {
		t.Fatal("expected allow")
	}
	if restrictionsAllow([]string{"get"}, []string{"field:gw.n.*"}, "update", "field:gw.n.s.f") {
		t.Fatal("expected deny action")
	}
}

// Full evaluation path: Allow mysensor.* tree + Deny node:mysensor.1
func TestDenyNodeDoesNotBlockSiblingFields(t *testing.T) {
	sts := []policyTY.Statement{
		{
			Effect:  policyTY.EffectAllow,
			Actions: []string{"*"},
			Resources: []string{
				"gateway:mysensor",
				"node:mysensor.*",
				"source:mysensor.*",
				"field:mysensor.*",
				"metric:mysensor.*",
			},
		},
		{
			Effect:    policyTY.EffectDeny,
			Actions:   []string{"*"},
			Resources: []string{"node:mysensor.1"},
		},
	}

	// siblings must remain allowed
	for _, res := range []string{
		"node:mysensor.2",
		"source:mysensor.2.s1",
		"field:mysensor.2.s1.temp",
		"metric:mysensor.2.s1.temp",
	} {
		if !statementsAllow(sts, "get", res) {
			t.Errorf("expected allow get %s", res)
		}
		if !statementsAllow(sts, "list", res) {
			t.Errorf("expected allow list %s", res)
		}
	}

	// under denied node: blocked
	for _, res := range []string{
		"node:mysensor.1",
		"source:mysensor.1.s1",
		"field:mysensor.1.s1.temp",
		"metric:mysensor.1.s1.temp",
	} {
		if statementsAllow(sts, "get", res) {
			t.Errorf("expected deny get %s", res)
		}
	}
}

func TestDenyNodeWithReadwriteStyleBareKinds(t *testing.T) {
	// like built-in readwrite + deny one node
	sts := []policyTY.Statement{
		{
			Effect:  policyTY.EffectAllow,
			Actions: []string{"*"},
			Resources: []string{
				"gateway", "node", "source", "field", "metric",
			},
		},
		{
			Effect:    policyTY.EffectDeny,
			Actions:   []string{"*"},
			Resources: []string{"node:mysensor.1"},
		},
	}
	if !statementsAllow(sts, "get", "field:mysensor.2.s.f") {
		t.Fatal("bare kind allow + node deny must still allow other node fields")
	}
	if statementsAllow(sts, "get", "field:mysensor.1.s.f") {
		t.Fatal("must deny fields under denied node")
	}
	if !statementsAllow(sts, "get", "source:mysensor.2.s") {
		t.Fatal("must allow other sources")
	}
	if statementsAllow(sts, "get", "source:mysensor.1.s") {
		t.Fatal("must deny sources under denied node")
	}
}

func TestMatchResourceDenyCascadeDoesNotMatchUnrelated(t *testing.T) {
	// kind-wide node deny should cascade to all fields
	if !MatchResourceDenyCascade("node", "field:a.b.c.d") {
		t.Fatal("kind-wide node deny should cascade")
	}
	if !MatchResourceDenyCascade("node:*", "field:a.b.c.d") {
		t.Fatal("node:* deny should cascade")
	}
	// named node must not cascade to unrelated
	if MatchResourceDenyCascade("node:mysensor.1", "field:other.1.s.f") {
		t.Fatal("must not cascade across gateways")
	}
	if MatchResourceDenyCascade("node:mysensor.1", "field:mysensor.2.s.f") {
		t.Fatal("must not cascade to sibling node fields")
	}
}
