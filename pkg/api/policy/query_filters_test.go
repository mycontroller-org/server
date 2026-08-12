package policy

import (
	"testing"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

func TestPatternToFilterGroupFieldExact(t *testing.T) {
	g := patternToFilterGroup(policyTY.ResourceField, "home-gw.living-room.dht.temp")
	if len(g) != 4 {
		t.Fatalf("want 4 filters, got %d: %+v", len(g), g)
	}
	want := map[string]string{
		types.KeyGatewayID: "home-gw",
		types.KeyNodeID:    "living-room",
		types.KeySourceID:  "dht",
		types.KeyFieldID:   "temp",
	}
	for _, f := range g {
		if f.Operator != storageTY.OperatorEqual {
			t.Fatalf("op: %s", f.Operator)
		}
		if want[f.Key] != f.Value {
			t.Fatalf("key %s: got %v want %v", f.Key, f.Value, want[f.Key])
		}
	}
}

func TestPatternToFilterGroupFieldPrefix(t *testing.T) {
	g := patternToFilterGroup(policyTY.ResourceField, "home-gw.living-room.*")
	if len(g) != 2 {
		t.Fatalf("want GatewayID+NodeID only, got %+v", g)
	}
	if g[0].Key != types.KeyGatewayID || g[0].Value != "home-gw" {
		t.Fatalf("gateway: %+v", g[0])
	}
	if g[1].Key != types.KeyNodeID || g[1].Value != "living-room" {
		t.Fatalf("node: %+v", g[1])
	}
}

func TestPatternToFilterGroupGateway(t *testing.T) {
	g := patternToFilterGroup(policyTY.ResourceGateway, "home-gw")
	if len(g) != 1 || g[0].Key != types.KeyID || g[0].Value != "home-gw" {
		t.Fatalf("got %+v", g)
	}
}

func TestPatternToFilterGroupGatewayPrefixIsSegmentSafe(t *testing.T) {
	g := patternToFilterGroup(policyTY.ResourceGateway, "home.*")
	if len(g) != 1 || g[0].Key != types.KeyID {
		t.Fatalf("got %+v", g)
	}
	if g[0].Operator != storageTY.OperatorRegexCaseSensitive {
		t.Fatalf("op: %s", g[0].Operator)
	}
	if g[0].Value != `^home($|\.)` {
		t.Fatalf("regex: %v", g[0].Value)
	}
}

func TestHierarchicalExcludeFilterNodeExact(t *testing.T) {
	// Deny node:mysensor.1 → NOR[GatewayID=mysensor, NodeID=1]
	f := hierarchicalExcludeFilter(policyTY.ResourceNode, "mysensor.1")
	if f == nil {
		t.Fatal("expected filter")
	}
	if f.Operator != storageTY.OperatorNor {
		t.Fatalf("want Nor, got %s %+v", f.Operator, f)
	}
	pos, ok := f.Value.([]storageTY.Filter)
	if !ok || len(pos) != 2 {
		t.Fatalf("want positive AND group of 2, got %#v", f.Value)
	}
	if pos[0].Key != types.KeyGatewayID || pos[0].Operator != storageTY.OperatorEqual || pos[0].Value != "mysensor" {
		t.Fatalf("gateway: %+v", pos[0])
	}
	if pos[1].Key != types.KeyNodeID || pos[1].Operator != storageTY.OperatorEqual || pos[1].Value != "1" {
		t.Fatalf("node: %+v", pos[1])
	}
}

func TestHierarchicalExcludeFilterGatewayPrefix(t *testing.T) {
	f := hierarchicalExcludeFilter(policyTY.ResourceNode, "mysensor.*")
	if f == nil {
		t.Fatal("expected filter")
	}
	if f.Operator != storageTY.OperatorNor {
		t.Fatalf("want Nor, got %+v", f)
	}
	pos, ok := f.Value.([]storageTY.Filter)
	if !ok || len(pos) != 1 {
		t.Fatalf("want single gateway equal, got %#v", f.Value)
	}
	if pos[0].Key != types.KeyGatewayID || pos[0].Value != "mysensor" {
		t.Fatalf("got %+v", pos[0])
	}
}

func TestAllowWildcardPlusDenyCombines(t *testing.T) {
	// Simulate StorageFiltersForList pattern list: allow mysensor.* + deny mysensor.1
	patterns := []string{"mysensor.*", "!mysensor.1"}
	allowPart, denyPart := splitAllowDenyPatterns(patterns)
	if len(allowPart) != 1 || allowPart[0] != "mysensor.*" {
		t.Fatalf("allow: %v", allowPart)
	}
	if len(denyPart) != 1 || denyPart[0] != "!mysensor.1" {
		t.Fatalf("deny: %v", denyPart)
	}
	allowG := patternToFilterGroup(policyTY.ResourceNode, "mysensor.*")
	if len(allowG) != 1 || allowG[0].Key != types.KeyGatewayID || allowG[0].Value != "mysensor" {
		t.Fatalf("allow filter: %+v", allowG)
	}
	excl := hierarchicalExcludeFilter(policyTY.ResourceNode, "mysensor.1")
	if excl == nil || excl.Operator != storageTY.OperatorNor {
		t.Fatalf("exclude: %+v", excl)
	}
}

// Deny node:mysensor.1 when listing fields → NOR path on field keys
func TestHierarchicalExcludeFilterFieldUnderNode(t *testing.T) {
	f := hierarchicalExcludeFilter(policyTY.ResourceField, "mysensor.1")
	if f == nil {
		t.Fatal("expected filter")
	}
	if f.Operator != storageTY.OperatorNor {
		t.Fatalf("want Nor, got %+v", f)
	}
	pos, ok := f.Value.([]storageTY.Filter)
	if !ok || len(pos) != 2 {
		t.Fatalf("want GatewayID+NodeID equal group, got %#v", f.Value)
	}
	if pos[0].Key != types.KeyGatewayID || pos[0].Value != "mysensor" {
		t.Fatalf("gateway: %+v", pos[0])
	}
	if pos[1].Key != types.KeyNodeID || pos[1].Value != "1" {
		t.Fatalf("node: %+v", pos[1])
	}
}

func TestItemAllowedForListDenyExclusion(t *testing.T) {
	// allow mysensor.* fields, deny under node 1
	patterns := []string{"mysensor.*", "!mysensor.1"}
	okField := fieldTY.Field{GatewayID: "mysensor", NodeID: "2", SourceID: "s", FieldID: "t"}
	badField := fieldTY.Field{GatewayID: "mysensor", NodeID: "1", SourceID: "s", FieldID: "t"}
	if !ItemAllowedForList(policyTY.ResourceField, patterns, okField) {
		t.Fatal("expected field under node 2 allowed")
	}
	if ItemAllowedForList(policyTY.ResourceField, patterns, badField) {
		t.Fatal("expected field under node 1 denied")
	}
}
