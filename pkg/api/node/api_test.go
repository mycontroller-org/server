package node

import (
	"testing"
	"time"
)

func TestParseOTATime(t *testing.T) {
	now := time.Date(2026, 9, 6, 2, 34, 48, 808783163, time.FixedZone("IST", 5*3600+30*60))
	if got, ok := parseOTATime(now); !ok || !got.Equal(now) {
		t.Fatalf("time.Time: ok=%v got=%v", ok, got)
	}
	s := "2026-09-06T02:34:48.808783163+05:30"
	got, ok := parseOTATime(s)
	if !ok {
		t.Fatal("RFC3339Nano string")
	}
	if !got.Equal(now) {
		t.Fatalf("parsed %v want %v", got, now)
	}
	if _, ok := parseOTATime(""); ok {
		t.Fatal("empty string")
	}
	if _, ok := parseOTATime(nil); ok {
		t.Fatal("nil")
	}
}

func TestFormatOTADuration(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{40 * time.Second, "40s"},
		{100 * time.Second, "1m40s"},
		{4*time.Minute + 37*time.Second, "4m37s"},
		{time.Hour + 2*time.Minute + 3*time.Second, "1h2m3s"},
		{0, "0s"},
		{-time.Second, "0s"},
	}
	for _, tc := range cases {
		if got := formatOTADuration(tc.d); got != tc.want {
			t.Fatalf("%v: got %q want %q", tc.d, got, tc.want)
		}
	}
}
