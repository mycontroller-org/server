package upgrade

import (
	"testing"

	"go.uber.org/zap"
)

func TestNeedsUserEnabledRestore(t *testing.T) {
	logger := zap.NewNop()
	cases := []struct {
		name          string
		backupVersion string
		lastUpgrade   string
		want          bool
	}{
		{name: "devel with last upgrade 2.2.0-2", backupVersion: "2.3.0-devel", lastUpgrade: "2.2.0-2", want: true},
		{name: "after 2.3.0-1 applied", backupVersion: "2.3.0-devel", lastUpgrade: "2.3.0-1", want: false},
		{name: "released 2.2.0 no lastUpgrade", backupVersion: "2.2.0", lastUpgrade: "", want: true},
		{name: "released 2.3.0 no lastUpgrade", backupVersion: "2.3.0", lastUpgrade: "", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := needsUserEnabledRestore(logger, tc.backupVersion, tc.lastUpgrade)
			if got != tc.want {
				t.Fatalf("needsUserEnabledRestore(%q, %q)=%v, want %v", tc.backupVersion, tc.lastUpgrade, got, tc.want)
			}
		})
	}
}
