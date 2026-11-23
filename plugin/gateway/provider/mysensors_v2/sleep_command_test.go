package mysensors

import (
	"testing"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
)

// resetPendingSleepStore clears the store to isolate tests.
// concurrency.Store does not expose Reset(), so we iterate over the keys.
func resetPendingSleepStore() {
	for _, k := range pendingSleepStore.Keys() {
		pendingSleepStore.Remove(k)
	}
}

func TestSleepCommandIsSleepingNode(t *testing.T) {
	n := &nodeTY.Node{Labels: cmap.CustomStringMap{}}
	if isSleepingNode(n) {
		t.Fatal("empty labels must not be sleeping")
	}
	n.Labels.Set(LabelSmartSleepNode, "true")
	if !isSleepingNode(n) {
		t.Fatal("smart_sleep_node=true must be sleeping")
	}
}

func TestSleepCommandRecentlySeen(t *testing.T) {
	n := &nodeTY.Node{LastSeen: time.Now()}
	if !isRecentlySeen(n, sleepRecentlySeen) {
		t.Fatal("just seen must be recent")
	}
	n.LastSeen = time.Now().Add(-1 * time.Minute)
	if isRecentlySeen(n, sleepRecentlySeen) {
		t.Fatal("1 min ago must not be recent for 30s threshold")
	}
}

func TestPendingSleepCount(t *testing.T) {
	resetPendingSleepStore()
	if pendingSleepCount() != 0 {
		t.Fatal("must start empty")
	}
	node := &nodeTY.Node{GatewayID: "gw", NodeID: "1"}
	p := &Provider{}
	p.queueSleepCommand(node, "206", 50, "req-1")
	if pendingSleepCount() != 1 {
		t.Fatalf("want 1 got %d", pendingSleepCount())
	}
}

func TestPurgeExpiredSleepCommands(t *testing.T) {
	resetPendingSleepStore()
	key := "gw_1"
	pendingSleepStore.Add(key, pendingSleepCommand{
		SensorID: "206",
		Percent:  75,
		QueuedAt: time.Now().Add(-3 * time.Hour),
	})
	purgeExpiredSleepCommands()
	if pendingSleepCount() != 0 {
		t.Fatalf("expired command must be purged")
	}
}

// --- Additional tests ---

func TestParseSleepDurationPayload(t *testing.T) {
	tests := []struct {
		in      string
		wantID  string
		wantPct uint8
		wantErr bool
	}{
		{"32", "", 32, false},
		{"0", "", 0, false},
		{"100", "", 100, false},
		{"206:32", "206", 32, false},
		{" 206 : 32 ", "206", 32, false},
		{"206:100", "206", 100, false},
		{"", "", 0, true},
		{"101", "", 0, true},
		{"-1", "", 0, true},
		{"foo", "", 0, true},
		{"206:foo", "", 0, true},
		{"206:200", "", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			id, pct, err := parseSleepDurationPayload(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tt.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.in, err)
			}
			if id != tt.wantID || pct != tt.wantPct {
				t.Fatalf("got (%q,%d), want (%q,%d)", id, pct, tt.wantID, tt.wantPct)
			}
		})
	}
}

func TestPendingSleepCommandKeepsSensorID(t *testing.T) {
	resetPendingSleepStore()
	node := &nodeTY.Node{GatewayID: "gw", NodeID: "1"}
	p := &Provider{}
	p.queueSleepCommand(node, "206", 50, "req-1")

	v := pendingSleepStore.Get("gw_1")
	cmd, ok := v.(pendingSleepCommand)
	if !ok {
		t.Fatal("expected pendingSleepCommand")
	}
	if cmd.SensorID != "206" || cmd.Percent != 50 {
		t.Fatalf("got SensorID=%q Percent=%d", cmd.SensorID, cmd.Percent)
	}
}
