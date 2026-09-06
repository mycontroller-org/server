package policy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	svcTokenTY "github.com/mycontroller-org/server/v2/pkg/types/service_token"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
)

func TestResourceFromQuickID(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"field:mysensor.1.3.V_VAR1", "field:mysensor.1.3.V_VAR1"},
		{"node:mysensor.2", "node:mysensor.2"},
		{"source:mysensor.2.4", "source:mysensor.2.4"},
		{"gateway:mysensor", "gateway:mysensor"},
		{"task:night-mode", "task:night-mode"},
	}
	for _, c := range cases {
		got, err := ResourceFromQuickID(c.in)
		if err != nil {
			t.Fatalf("%s: %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("%s: got %q want %q", c.in, got, c.want)
		}
	}
}

func TestAuthorizeQuickIDRequest_DeniesNode1Field(t *testing.T) {
	p := policyTY.Policy{
		ID: "test",
		Statements: []policyTY.Statement{
			{
				Effect:  policyTY.EffectAllow,
				Actions: []string{"get", "list"},
				Resources: []string{
					"gateway:mysensor",
					"node:mysensor.*",
					"quickid",
				},
			},
			{
				Effect:    policyTY.EffectDeny,
				Actions:   []string{"get", "list"},
				Resources: []string{"node:mysensor.1"},
			},
		},
	}
	c := newCache()
	c.PutUser(&userTY.User{ID: "u1", Policies: []string{"test"}})
	cp := p
	c.PutPolicy(&cp)
	c.setLoaders(
		func(id string) (*userTY.User, error) { return nil, ErrUserNotFound },
		func(id string) (*policyTY.Policy, error) { return nil, ErrUserNotFound },
		func(id string) (*svcTokenTY.ServiceToken, error) { return nil, ErrTokenNotFound },
		func() ([]policyTY.Policy, error) { return nil, nil },
	)
	a := &API{cache: c}
	sub := Subject{UserID: "u1"}

	// bare quickid without ids: allowed (coarse)
	r0 := httptest.NewRequest(http.MethodGet, "/api/quickid", nil)
	if err := a.AuthorizeQuickIDRequest(sub, r0); err != nil {
		t.Fatalf("bare quickid: %v", err)
	}

	// field under denied node
	r1 := httptest.NewRequest(http.MethodGet, "/api/quickid?id=field:mysensor.1.3.V_VAR1", nil)
	if err := a.AuthorizeQuickIDRequest(sub, r1); err == nil {
		t.Fatal("expected deny field under node 1 via quickid")
	}

	// field under allowed node
	r2 := httptest.NewRequest(http.MethodGet, "/api/quickid?id=field:mysensor.2.4.V_STATUS", nil)
	if err := a.AuthorizeQuickIDRequest(sub, r2); err != nil {
		t.Fatalf("expected allow field under node 2: %v", err)
	}

	// multi-id with one denied → fail whole request
	r3 := httptest.NewRequest(http.MethodGet, "/api/quickid?id=field:mysensor.2.4.V_STATUS&id=field:mysensor.1.3.V_VAR1", nil)
	if err := a.AuthorizeQuickIDRequest(sub, r3); err == nil {
		t.Fatal("expected deny when any id is denied")
	}
}
