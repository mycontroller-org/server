package policy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
)

func TestAuthorizeSleepingQueueRequestNamedGrant(t *testing.T) {
	a := apiWithPolicies(t, singleGatewayPolicy())
	subject := Subject{UserID: "u1"}

	ok := httptest.NewRequest(http.MethodGet, "/api/gateway-sleeping-queue?gatewayId=gw1", nil)
	if err := a.AuthorizeSleepingQueueRequest(subject, ok); err != nil {
		t.Fatalf("gw1: %v", err)
	}

	other := httptest.NewRequest(http.MethodGet, "/api/gateway-sleeping-queue?gatewayId=gw2", nil)
	if err := a.AuthorizeSleepingQueueRequest(subject, other); err == nil {
		t.Fatal("expected deny for gw2")
	}

	missing := httptest.NewRequest(http.MethodGet, "/api/gateway-sleeping-queue", nil)
	if err := a.AuthorizeSleepingQueueRequest(subject, missing); err == nil {
		t.Fatal("expected deny when gatewayId is missing")
	}

	clearOther := httptest.NewRequest(http.MethodGet, "/api/gateway-sleeping-queue/clear?gatewayId=gw2", nil)
	if err := a.AuthorizeSleepingQueueRequest(subject, clearOther); err == nil {
		t.Fatal("expected deny clear for gw2")
	}
}

func TestAuthorizeSleepingQueueRequestNodeTarget(t *testing.T) {
	a := apiWithPolicies(t, policyTY.Policy{
		ID: "one-node",
		Statements: []policyTY.Statement{{
			Effect:    policyTY.EffectAllow,
			Actions:   []string{policyTY.ActionGet, policyTY.ActionAction},
			Resources: []string{"node:gw1.n1"},
		}},
	})
	subject := Subject{UserID: "u1"}

	ok := httptest.NewRequest(http.MethodGet, "/api/gateway-sleeping-queue?gatewayId=gw1&nodeId=n1", nil)
	if err := a.AuthorizeSleepingQueueRequest(subject, ok); err != nil {
		t.Fatalf("node n1: %v", err)
	}

	other := httptest.NewRequest(http.MethodGet, "/api/gateway-sleeping-queue?gatewayId=gw1&nodeId=n2", nil)
	if err := a.AuthorizeSleepingQueueRequest(subject, other); err == nil {
		t.Fatal("expected deny for node n2")
	}
}
