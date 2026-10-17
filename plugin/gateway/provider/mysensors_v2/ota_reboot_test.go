package mysensors

import (
	"testing"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
)

func TestShouldRebootForOta(t *testing.T) {
	tests := []struct {
		name  string
		label string
		want  bool
	}{
		{"absent", "", false},
		{"true", "true", true},
		{"false", "false", false},
		{"TRUE", "TRUE", true},
		{"1", "1", true},
		{"0", "0", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labels := cmap.CustomStringMap{}
			if tt.label != "" {
				labels.Set(LabelBootloaderOtaReboot, tt.label)
			}
			n := &nodeTY.Node{Labels: labels}
			if got := shouldRebootForOta(n); got != tt.want {
				t.Errorf("shouldRebootForOta(%q) = %v, want %v", tt.label, got, tt.want)
			}
		})
	}

	if shouldRebootForOta(nil) {
		t.Fatal("nil node must be false")
	}
}

func TestIsSleepingNode(t *testing.T) {
	n := &nodeTY.Node{Labels: cmap.CustomStringMap{}}
	if isSleepingNode(n) {
		t.Fatal("empty labels must not be sleeping")
	}
	n.Labels.Set(LabelSmartSleepNode, "true")
	if !isSleepingNode(n) {
		t.Fatal("smart_sleep_node=true must be sleeping")
	}
	n.Labels.Set(LabelSmartSleepNode, "false")
	if isSleepingNode(n) {
		t.Fatal("smart_sleep_node=false must not be sleeping")
	}
}

func TestIsRecentlySeen(t *testing.T) {
	n := &nodeTY.Node{LastSeen: time.Now()}
	if !isRecentlySeen(n, 30*time.Second) {
		t.Fatal("just seen must be recent")
	}
	n.LastSeen = time.Now().Add(-1 * time.Minute)
	if isRecentlySeen(n, 30*time.Second) {
		t.Fatal("1 min ago must not be recent for 30s threshold")
	}
}

func TestMaxSleepDuration(t *testing.T) {
	n := &nodeTY.Node{Labels: cmap.CustomStringMap{}}
	if got := maxSleepDuration(n); got != 5*time.Minute {
		t.Fatalf("default: %v", got)
	}
	n.Labels.Set(FieldSleepDuration, "280000") // 280000 ms = 280 s
	if got := maxSleepDuration(n); got < 4*time.Minute || got > 5*time.Minute {
		t.Fatalf("280000ms: %v", got)
	}
	n.Labels.Set(FieldSleepDuration, "10000") // 10 s -> clamped to 1 min
	if got := maxSleepDuration(n); got != time.Minute {
		t.Fatalf("10s clamp: %v", got)
	}
	n.Labels.Set(FieldSleepDuration, "7200000") // 2 h -> clamped to 1 h
	if got := maxSleepDuration(n); got != time.Hour {
		t.Fatalf("2h clamp: %v", got)
	}
}

func TestMaybeSendRebootDecision(t *testing.T) {
	tests := []struct {
		name          string
		labels        cmap.CustomStringMap
		wantNeedsBoot bool
	}{
		{
			name:          "no labels, no detection",
			labels:        cmap.CustomStringMap{},
			wantNeedsBoot: false,
		},
		{
			name:          "explicit opt-in",
			labels:        cmap.CustomStringMap{LabelBootloaderOtaReboot: "true"},
			wantNeedsBoot: true,
		},
		{
			name:          "detected MYSBootloader 1.4",
			labels:        cmap.CustomStringMap{LabelBootloaderKind: BootloaderKindMYSBootloader14},
			wantNeedsBoot: true,
		},
		{
			name:          "detected MYSBootloader 1.0",
			labels:        cmap.CustomStringMap{LabelBootloaderKind: BootloaderKindMYSBootloader10},
			wantNeedsBoot: true,
		},
		{
			name:          "detected protocol 3.1, no opt-in",
			labels:        cmap.CustomStringMap{LabelBootloaderKind: BootloaderKindProtocol31},
			wantNeedsBoot: false,
		},
		{
			name: "protocol 3.1 with explicit opt-in",
			labels: cmap.CustomStringMap{
				LabelBootloaderKind:      BootloaderKindProtocol31,
				LabelBootloaderOtaReboot: "true",
			},
			wantNeedsBoot: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &nodeTY.Node{Labels: tt.labels}
			got := shouldRebootForOta(n) || isMYSBootloaderKind(bootloaderKindForNode(n))
			if got != tt.wantNeedsBoot {
				t.Errorf("needsReboot = %v, want %v", got, tt.wantNeedsBoot)
			}
		})
	}
}

func TestPendingRebootCount(t *testing.T) {
	// Nettoyer l'état entre les tests (le store est global au package).
	for _, k := range otaPendingRebootStore.Keys() {
		otaPendingRebootStore.Remove(k)
	}
	defer func() {
		for _, k := range otaPendingRebootStore.Keys() {
			otaPendingRebootStore.Remove(k)
		}
	}()

	if n := pendingRebootCount(); n != 0 {
		t.Fatalf("expected 0 pending reboots, got %d", n)
	}

	otaPendingRebootStore.Add("gw1.node1", time.Now())
	otaPendingRebootStore.Add("gw1.node2", time.Now())

	if n := pendingRebootCount(); n != 2 {
		t.Fatalf("expected 2 pending reboots, got %d", n)
	}

	otaPendingRebootStore.Remove("gw1.node1")

	if n := pendingRebootCount(); n != 1 {
		t.Fatalf("expected 1 pending reboot, got %d", n)
	}

	otaPendingRebootStore.Remove("gw1.node2")

	if n := pendingRebootCount(); n != 0 {
		t.Fatalf("expected 0 pending reboots after cleanup, got %d", n)
	}
}
