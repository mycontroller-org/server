package policy

import (
	"testing"

	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	svcTokenTY "github.com/mycontroller-org/server/v2/pkg/types/service_token"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
)

// tokenScopeAPI builds an API with one user, one policy and one service token in cache.
func tokenScopeAPI(t *testing.T, p policyTY.Policy, tokenResources []string) *API {
	t.Helper()
	c := newCache()
	c.PutUser(&userTY.User{ID: "u1", Username: "u", Policies: []string{p.ID}})
	cp := p
	c.PutPolicy(&cp)
	c.PutToken(&svcTokenTY.ServiceToken{
		ID:          "t-entity",
		UserID:      "u1",
		NeverExpire: true,
		Token:       svcTokenTY.Token{ID: "t1"},
		Resources:   tokenResources,
	})
	c.setLoaders(
		func(id string) (*userTY.User, error) { return nil, ErrUserNotFound },
		func(id string) (*policyTY.Policy, error) { return nil, ErrNotFound },
		func(tokenID string) (*svcTokenTY.ServiceToken, error) { return nil, ErrTokenNotFound },
		func() ([]policyTY.Policy, error) { return nil, nil },
	)
	return &API{cache: c}
}

// A service token must only be able to narrow the user's scope, never widen it.
func TestResourceNamesForList_TokenCanOnlyNarrow(t *testing.T) {
	userPolicy := policyTY.Policy{
		ID: "p1",
		Statements: []policyTY.Statement{{
			Effect:    policyTY.EffectAllow,
			Actions:   []string{"*"},
			Resources: []string{"field:mysensor.*"},
		}},
	}

	cases := []struct {
		name           string
		tokenResources []string
		wantAllow      []string
		wantDenied     bool
	}{
		{
			name:           "token narrows to one field",
			tokenResources: []string{"field:mysensor.1.s.temp"},
			wantAllow:      []string{"mysensor.1.s.temp"},
		},
		{
			name:           "token names a gateway outside the user scope",
			tokenResources: []string{"gateway:other-gw"},
			wantAllow:      nil, // list api reachable, but no overlap -> no rows
		},
		{
			name:           "token names an unrelated kind only",
			tokenResources: []string{"task:night-mode"},
			wantDenied:     true, // token does not reach this kind at all
		},
		{
			name:           "token parent gateway cascades to fields",
			tokenResources: []string{"gateway:mysensor"},
			wantAllow:      []string{"mysensor"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := tokenScopeAPI(t, userPolicy, tc.tokenResources)
			subject := Subject{UserID: "u1", ServiceTokenID: "t1"}

			unrestricted, patterns, err := a.ResourceNamesForList(subject, policyTY.ResourceField)
			if tc.wantDenied {
				if err == nil {
					t.Fatalf("expected access denied, got patterns %v", patterns)
				}
				return
			}
			if err != nil {
				t.Fatalf("ResourceNamesForList: %v", err)
			}
			if unrestricted {
				t.Fatal("token restricted subject must never be unrestricted")
			}
			allow, _ := splitAllowDenyPatterns(patterns)
			if len(allow) != len(tc.wantAllow) {
				t.Fatalf("allow patterns = %v, want %v", allow, tc.wantAllow)
			}
			for i := range tc.wantAllow {
				if allow[i] != tc.wantAllow[i] {
					t.Fatalf("allow patterns = %v, want %v", allow, tc.wantAllow)
				}
			}

			// an empty scope must produce a query that returns nothing,
			// not a query without scope filters
			_, filters, err := a.StorageFiltersForList(subject, policyTY.ResourceField)
			if err != nil {
				t.Fatalf("StorageFiltersForList: %v", err)
			}
			if len(filters) == 0 {
				t.Fatal("expected scope filters, got none (would list every row)")
			}
		})
	}
}

// Deny patterns must not survive alone when the token empties the allow scope:
// a deny-only pattern list means "everything except ..." downstream.
func TestResourceNamesForList_TokenEmptiesScopeDropsDenyOnly(t *testing.T) {
	userPolicy := policyTY.Policy{
		ID: "p1",
		Statements: []policyTY.Statement{
			{
				Effect:    policyTY.EffectAllow,
				Actions:   []string{"*"},
				Resources: []string{"field:mysensor.*"},
			},
			{
				Effect:    policyTY.EffectDeny,
				Actions:   []string{"*"},
				Resources: []string{"field:mysensor.1.s.f"},
			},
		},
	}
	// token reaches the field kind (gateway is an ancestor) but names a gateway
	// the user policy does not cover
	a := tokenScopeAPI(t, userPolicy, []string{"gateway:other-gw"})
	subject := Subject{UserID: "u1", ServiceTokenID: "t1"}

	_, patterns, err := a.ResourceNamesForList(subject, policyTY.ResourceField)
	if err != nil {
		t.Fatalf("ResourceNamesForList: %v", err)
	}
	allow, deny := splitAllowDenyPatterns(patterns)
	if len(allow) != 0 || len(deny) != 0 {
		t.Fatalf("expected empty scope, got allow=%v deny=%v", allow, deny)
	}
}
