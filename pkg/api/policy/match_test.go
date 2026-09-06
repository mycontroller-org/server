package policy

import "testing"

func TestMatchAction(t *testing.T) {
	cases := []struct {
		pattern, action string
		want            bool
	}{
		{"*", "get", true},
		{"get", "get", true},
		{"get", "list", false},
		{"", "get", false},
	}
	for _, c := range cases {
		if got := MatchAction(c.pattern, c.action); got != c.want {
			t.Errorf("MatchAction(%q,%q)=%v want %v", c.pattern, c.action, got, c.want)
		}
	}
}

func TestMatchResource(t *testing.T) {
	cases := []struct {
		pattern, resource string
		want              bool
	}{
		{"*", "field:a.b.c.d", true},
		{"field", "field:a.b.c.d", true},
		{"field:*", "field:a.b.c.d", true},
		{"field:a.b.c.d", "field:a.b.c.d", true},
		{"field:a.b.c.d", "field:a.b.c.other", false},
		{"field:home-gw.living-room.*", "field:home-gw.living-room.dht.temp", true},
		{"field:home-gw.living-room.*", "field:home-gw.living-room", true},
		{"field:home-gw.living-room.*", "field:home-gw.kitchen.dht.temp", false},
		{"node:home-gw.*", "node:home-gw.living-room", true},
		{"gateway:home-gw", "gateway:home-gw", true},
		{"gateway:home-gw", "node:home-gw.x", false},
		{"task:night-mode", "task:night-mode", true},
		// collection (list) checks use kind without name
		{"field:home-gw.living-room.*", "field", true},
		{"gateway:home-gw", "gateway", true},
		{"node:x", "field", false},
	}
	for _, c := range cases {
		if got := MatchResource(c.pattern, c.resource); got != c.want {
			t.Errorf("MatchResource(%q,%q)=%v want %v", c.pattern, c.resource, got, c.want)
		}
	}
}

func TestMatchResourceDenyCascade(t *testing.T) {
	cases := []struct {
		pattern, resource string
		want              bool
	}{
		// same-kind still works
		{"node:mysensor.1", "node:mysensor.1", true},
		{"node:mysensor.1", "node:mysensor.2", false},
		// node deny → field/source/metric under path
		{"node:mysensor.1", "field:mysensor.1.s1.temp", true},
		{"node:mysensor.1", "source:mysensor.1.s1", true},
		{"node:mysensor.1", "metric:mysensor.1.s1.temp", true},
		{"node:mysensor.1", "field:mysensor.2.s1.temp", false},
		// gateway deny → children
		{"gateway:mysensor", "node:mysensor.1", true},
		{"gateway:mysensor", "field:mysensor.1.s.f", true},
		{"gateway:mysensor", "field:other.1.s.f", false},
		// source deny → field
		{"source:mysensor.1.s1", "field:mysensor.1.s1.temp", true},
		{"source:mysensor.1.s1", "field:mysensor.1.other.temp", false},
		// Allow-style MatchResource must stay false for cross-kind
		// (cascade is Deny-only API)
		{"node:mysensor.1", "task:x", false},
		// kind-wide parent deny
		{"gateway:*", "field:a.b.c.d", true},
		{"node", "field:a.b.c.d", true},
		// named parent does not block collection resource
		{"node:mysensor.1", "field", false},
	}
	for _, c := range cases {
		if got := MatchResourceDenyCascade(c.pattern, c.resource); got != c.want {
			t.Errorf("MatchResourceDenyCascade(%q,%q)=%v want %v", c.pattern, c.resource, got, c.want)
		}
	}
	// Ensure plain MatchResource still does not cascade
	if MatchResource("node:mysensor.1", "field:mysensor.1.s.f") {
		t.Fatal("MatchResource must not cascade across kinds")
	}
}
