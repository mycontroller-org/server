package policy

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	svcTokenTY "github.com/mycontroller-org/server/v2/pkg/types/service_token"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
)

func apiWithPolicies(t *testing.T, policies ...policyTY.Policy) *API {
	t.Helper()
	c := newCache()
	ids := make([]string, 0, len(policies))
	for i := range policies {
		cp := policies[i]
		c.PutPolicy(&cp)
		ids = append(ids, cp.ID)
	}
	c.PutUser(&userTY.User{ID: "u1", Username: "u", Policies: ids})
	c.setLoaders(
		func(id string) (*userTY.User, error) { return nil, ErrUserNotFound },
		func(id string) (*policyTY.Policy, error) { return nil, ErrNotFound },
		func(tokenID string) (*svcTokenTY.ServiceToken, error) { return nil, ErrTokenNotFound },
		func() ([]policyTY.Policy, error) { return nil, nil },
	)
	return &API{cache: c}
}

// one named gateway, full verbs on it
func singleGatewayPolicy() policyTY.Policy {
	return policyTY.Policy{
		ID: "one-gateway",
		Statements: []policyTY.Statement{{
			Effect:    policyTY.EffectAllow,
			Actions:   []string{"*"},
			Resources: []string{"gateway:gw1"},
		}},
	}
}

func writeRequest(t *testing.T, method, path, body string) *http.Request {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	return r
}

// A grant on one named object must not authorize writes to its siblings, even
// though the collection endpoint itself is reachable.
func TestAuthorizeBodyTargets_NamedGrantDoesNotEscapeToSiblings(t *testing.T) {
	a := apiWithPolicies(t, singleGatewayPolicy())
	subject := Subject{UserID: "u1"}

	cases := []struct {
		name    string
		method  string
		path    string
		body    string
		allowed bool
	}{
		{"update own gateway", http.MethodPost, "/api/gateway", `{"id":"gw1","name":"a"}`, true},
		{"update other gateway", http.MethodPost, "/api/gateway", `{"id":"gw2","name":"a"}`, false},
		{"enable own gateway", http.MethodPost, "/api/gateway/enable", `["gw1"]`, true},
		{"enable other gateway", http.MethodPost, "/api/gateway/enable", `["gw2"]`, false},
		{"enable own + other", http.MethodPost, "/api/gateway/enable", `["gw1","gw2"]`, false},
		{"delete own gateway", http.MethodDelete, "/api/gateway", `["gw1"]`, true},
		{"delete other gateway", http.MethodDelete, "/api/gateway", `["gw2"]`, false},
		// no target in the payload: needs a kind-wide grant, which this policy lacks
		{"create with generated id", http.MethodPost, "/api/gateway", `{"name":"a"}`, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := writeRequest(t, tc.method, tc.path, tc.body)
			access := MapRequest(r)

			// the coarse gate must stay permissive, or the object level check
			// would never run
			if err := a.Allowed(subject, access.Action, access.Resource); err != nil {
				t.Fatalf("collection gate denied the request: %v", err)
			}

			err := a.AuthorizeBodyTargets(subject, r, &access)
			if tc.allowed && err != nil {
				t.Fatalf("expected allowed, got %v", err)
			}
			if !tc.allowed && err == nil {
				t.Fatal("expected denied, got allowed")
			}

			// the handler must still be able to read the payload
			body := make([]byte, len(tc.body))
			n, _ := r.Body.Read(body)
			if string(body[:n]) != tc.body {
				t.Fatalf("body not restored: %q", string(body[:n]))
			}
		})
	}
}

// A kind-wide grant keeps working for every shape, including create.
func TestAuthorizeBodyTargets_KindWideGrant(t *testing.T) {
	a := apiWithPolicies(t, policyTY.Policy{
		ID: "all-gateways",
		Statements: []policyTY.Statement{{
			Effect:    policyTY.EffectAllow,
			Actions:   []string{"*"},
			Resources: []string{policyTY.ResourceGateway},
		}},
	})
	subject := Subject{UserID: "u1"}

	bodies := []string{`{"id":"gw2"}`, `{"name":"new"}`, `["gw1","gw2","gw3"]`}
	for _, body := range bodies {
		r := writeRequest(t, http.MethodPost, "/api/gateway", body)
		access := MapRequest(r)
		if err := a.AuthorizeBodyTargets(subject, r, &access); err != nil {
			t.Fatalf("body %s: expected allowed, got %v", body, err)
		}
	}
}

// Device tree payloads name their target through the gateway/node/source/field
// chain rather than the storage id.
func TestAuthorizeBodyTargets_DeviceTreePath(t *testing.T) {
	a := apiWithPolicies(t, policyTY.Policy{
		ID: "one-node",
		Statements: []policyTY.Statement{{
			Effect:    policyTY.EffectAllow,
			Actions:   []string{"*"},
			Resources: []string{"node:gw1.n1"},
		}},
	})
	subject := Subject{UserID: "u1"}

	allowed := writeRequest(t, http.MethodPost, "/api/field",
		`{"gatewayId":"gw1","nodeId":"n1","sourceId":"s1","fieldId":"temp"}`)
	access := MapRequest(allowed)
	if err := a.AuthorizeBodyTargets(subject, allowed, &access); err != nil {
		t.Fatalf("field under the allowed node: %v", err)
	}

	denied := writeRequest(t, http.MethodPost, "/api/field",
		`{"gatewayId":"gw1","nodeId":"n2","sourceId":"s1","fieldId":"temp"}`)
	access = MapRequest(denied)
	if err := a.AuthorizeBodyTargets(subject, denied, &access); err == nil {
		t.Fatal("field under another node must be denied")
	}
}

// An explicit deny on one object survives the body level check.
func TestAuthorizeBodyTargets_DenyWins(t *testing.T) {
	a := apiWithPolicies(t, policyTY.Policy{
		ID: "all-but-one",
		Statements: []policyTY.Statement{
			{Effect: policyTY.EffectAllow, Actions: []string{"*"}, Resources: []string{policyTY.ResourceGateway}},
			{Effect: policyTY.EffectDeny, Actions: []string{"*"}, Resources: []string{"gateway:critical"}},
		},
	})
	subject := Subject{UserID: "u1"}

	r := writeRequest(t, http.MethodDelete, "/api/gateway", `["gw1","critical"]`)
	access := MapRequest(r)
	if err := a.AuthorizeBodyTargets(subject, r, &access); err == nil {
		t.Fatal("deny on gateway:critical must block the bulk delete")
	}
}

// Uploads name their target in the path, so their body is never buffered.
func TestAuthorizeBodyTargets_SkipsPathNamedUpload(t *testing.T) {
	a := apiWithPolicies(t, singleGatewayPolicy())
	subject := Subject{UserID: "u1"}

	r := httptest.NewRequest(http.MethodPost, "/api/firmware/upload/fw1", strings.NewReader("binary-blob"))
	r.Header.Set("Content-Type", "multipart/form-data; boundary=x")
	access := MapRequest(r)
	if access.Name != "fw1" {
		t.Fatalf("expected the path to name the target, got %q", access.Name)
	}
	if err := a.AuthorizeBodyTargets(subject, r, &access); err != nil {
		t.Fatalf("upload must not be body checked: %v", err)
	}
}

// Content-Type must not be a way around the object level check: handlers decode
// json whatever the header claims.
func TestAuthorizeBodyTargets_ContentTypeCannotBypass(t *testing.T) {
	a := apiWithPolicies(t, singleGatewayPolicy())
	subject := Subject{UserID: "u1"}

	for _, contentType := range []string{"", "text/plain", "application/x-www-form-urlencoded", "application/json"} {
		r := httptest.NewRequest(http.MethodPost, "/api/gateway", strings.NewReader(`{"id":"gw2"}`))
		if contentType != "" {
			r.Header.Set("Content-Type", contentType)
		}
		access := MapRequest(r)
		if err := a.AuthorizeBodyTargets(subject, r, &access); err == nil {
			t.Fatalf("content-type %q bypassed the object level check", contentType)
		}
	}
}

// A policy naming a single user must not be able to create or edit other users.
func TestAuthorizeBodyTargets_UserEscalation(t *testing.T) {
	a := apiWithPolicies(t, policyTY.Policy{
		ID: "self-service",
		Statements: []policyTY.Statement{{
			Effect:    policyTY.EffectAllow,
			Actions:   []string{policyTY.ActionGet, policyTY.ActionUpdate},
			Resources: []string{"user:u1"},
		}},
	})
	subject := Subject{UserID: "u1"}

	for _, body := range []string{
		`{"id":"admin-user-id","policies":["admin"]}`, // hijack another user
		`{"username":"new","policies":["admin"]}`,     // create a new admin
	} {
		r := writeRequest(t, http.MethodPost, "/api/user", body)
		access := MapRequest(r)
		if err := a.AuthorizeBodyTargets(subject, r, &access); err == nil {
			t.Fatalf("expected denied for body %s", body)
		}
	}

	// its own record is still writable
	r := writeRequest(t, http.MethodPost, "/api/user", `{"id":"u1","fullName":"me"}`)
	access := MapRequest(r)
	if err := a.AuthorizeBodyTargets(subject, r, &access); err != nil {
		t.Fatalf("own user record: %v", err)
	}

	for _, body := range []string{
		`{"id":"u1","policies":["admin"]}`,
		`{"id":"u1","disabled":true}`,
	} {
		r := writeRequest(t, http.MethodPost, "/api/user", body)
		access := MapRequest(r)
		if err := a.AuthorizeBodyTargets(subject, r, &access); err == nil {
			t.Fatalf("expected privileged field denied for body %s", body)
		}
	}
}
