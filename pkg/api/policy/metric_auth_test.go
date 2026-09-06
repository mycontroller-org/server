package policy

import (
	"testing"

	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
)

func TestMetricPathMatching(t *testing.T) {
	// hierarchical metric patterns reuse MatchResource
	cases := []struct {
		pattern, resource string
		want              bool
	}{
		{"metric:mysensor.*", "metric:mysensor.1.s.temp", true},
		{"metric:mysensor.1.*", "metric:mysensor.1.s.temp", true},
		{"metric:mysensor.1.s.*", "metric:mysensor.1.s.temp", true},
		{"metric:mysensor.1.s.temp", "metric:mysensor.1.s.temp", true},
		{"metric:mysensor.*", "metric:other.1.s.temp", false},
		{"metric", "metric:mysensor.1.s.temp", true}, // kind-only = all metrics
		{"metric:*", "metric:mysensor.1.s.temp", true},
		{"*", "metric:mysensor.1.s.temp", true},
	}
	for _, c := range cases {
		if got := MatchResource(c.pattern, c.resource); got != c.want {
			t.Errorf("MatchResource(%q,%q)=%v want %v", c.pattern, c.resource, got, c.want)
		}
	}
	_ = policyTY.ResourceMetric
}
