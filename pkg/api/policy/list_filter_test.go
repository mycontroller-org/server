package policy

import (
	"testing"

	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	gatewayTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
)

func TestBusinessName(t *testing.T) {
	if got := BusinessName(policyTY.ResourceGateway, gatewayTY.Config{ID: "home-gw"}); got != "home-gw" {
		t.Fatalf("gateway name: %q", got)
	}
	if got := BusinessName(policyTY.ResourceNode, nodeTY.Node{GatewayID: "home-gw", NodeID: "living-room"}); got != "home-gw.living-room" {
		t.Fatalf("node name: %q", got)
	}
	if got := BusinessName(policyTY.ResourceField, fieldTY.Field{
		GatewayID: "home-gw", NodeID: "living-room", SourceID: "dht", FieldID: "temp",
	}); got != "home-gw.living-room.dht.temp" {
		t.Fatalf("field name: %q", got)
	}
}

func TestItemAllowedForList(t *testing.T) {
	field := fieldTY.Field{GatewayID: "home-gw", NodeID: "living-room", SourceID: "dht", FieldID: "temp"}
	patterns := []string{"home-gw.living-room.*"}
	if !ItemAllowedForList(policyTY.ResourceField, patterns, field) {
		t.Fatal("expected field allowed")
	}
	other := fieldTY.Field{GatewayID: "home-gw", NodeID: "kitchen", SourceID: "dht", FieldID: "temp"}
	if ItemAllowedForList(policyTY.ResourceField, patterns, other) {
		t.Fatal("expected kitchen field denied")
	}
}
