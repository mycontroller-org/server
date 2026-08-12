package policy

import (
	"strings"
	"testing"

	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	svcTokenTY "github.com/mycontroller-org/server/v2/pkg/types/service_token"
	sourceTY "github.com/mycontroller-org/server/v2/pkg/types/source"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	filterUtils "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
)

func mockAPIWithPolicy(t *testing.T, userID string, p policyTY.Policy) *API {
	t.Helper()
	c := newCache()
	u := &userTY.User{ID: userID, Username: "u", Policies: []string{p.ID}}
	cp := p
	c.PutUser(u)
	c.PutPolicy(&cp)
	c.setLoaders(
		func(id string) (*userTY.User, error) { return nil, ErrUserNotFound },
		func(id string) (*policyTY.Policy, error) { return nil, ErrUserNotFound },
		func(id string) (*svcTokenTY.ServiceToken, error) { return nil, ErrTokenNotFound },
		func() ([]policyTY.Policy, error) { return nil, nil },
	)
	return &API{cache: c}
}

func TestResourceNamesForList_DenyNodeExpandsToFieldExclude(t *testing.T) {
	p := policyTY.Policy{
		ID: "p1",
		Statements: []policyTY.Statement{
			{
				Effect:  policyTY.EffectAllow,
				Actions: []string{"*"},
				Resources: []string{
					"gateway:mysensor",
					"node:mysensor.*",
					"source:mysensor.*",
					"field:mysensor.*",
				},
			},
			{
				Effect:    policyTY.EffectDeny,
				Actions:   []string{"*"},
				Resources: []string{"node:mysensor.1"},
			},
		},
	}
	a := mockAPIWithPolicy(t, "u1", p)
	sub := Subject{UserID: "u1"}

	for _, kind := range []string{policyTY.ResourceField, policyTY.ResourceSource, policyTY.ResourceNode} {
		unrestricted, patterns, err := a.ResourceNamesForList(sub, kind)
		if err != nil {
			t.Fatalf("%s: ResourceNamesForList err: %v", kind, err)
		}
		if unrestricted {
			t.Fatalf("%s: expected restricted list", kind)
		}
		t.Logf("%s patterns: %v", kind, patterns)
		allow, deny := splitAllowDenyPatterns(patterns)
		if len(deny) == 0 {
			t.Fatalf("%s: expected deny patterns, got allow=%v deny=%v", kind, allow, deny)
		}
		found := false
		for _, d := range deny {
			if strings.TrimPrefix(d, "!") == "mysensor.1" {
				found = true
			}
		}
		if !found {
			t.Fatalf("%s: expected !mysensor.1 in deny, got %v", kind, deny)
		}
	}
}

func TestStorageFiltersForList_SiblingFieldsRemain(t *testing.T) {
	p := policyTY.Policy{
		ID: "p1",
		Statements: []policyTY.Statement{
			{
				Effect:  policyTY.EffectAllow,
				Actions: []string{"*"},
				Resources: []string{
					"node:mysensor.*",
					"source:mysensor.*",
					"field:mysensor.*",
				},
			},
			{
				Effect:    policyTY.EffectDeny,
				Actions:   []string{"*"},
				Resources: []string{"node:mysensor.1"},
			},
		},
	}
	a := mockAPIWithPolicy(t, "u1", p)
	sub := Subject{UserID: "u1"}

	unrestricted, filters, err := a.StorageFiltersForList(sub, policyTY.ResourceField)
	if err != nil {
		t.Fatal(err)
	}
	if unrestricted {
		t.Fatal("expected restricted")
	}
	t.Logf("filters: %+v", filters)

	entities := []interface{}{
		&fieldTY.Field{ID: "a", GatewayID: "mysensor", NodeID: "1", SourceID: "s", FieldID: "t"},
		&fieldTY.Field{ID: "b", GatewayID: "mysensor", NodeID: "2", SourceID: "s", FieldID: "t"},
		&fieldTY.Field{ID: "c", GatewayID: "mysensor", NodeID: "3", SourceID: "s", FieldID: "t"},
	}
	matched := filterUtils.Filter(entities, filters, false)
	ids := []string{}
	for _, m := range matched {
		ids = append(ids, m.(*fieldTY.Field).NodeID)
	}
	t.Logf("matched nodeIDs: %v", ids)
	if len(matched) != 2 {
		t.Fatalf("want 2 sibling fields, got %d filters=%+v", len(matched), filters)
	}
	for _, m := range matched {
		if m.(*fieldTY.Field).NodeID == "1" {
			t.Fatal("denied node field should not match")
		}
	}

	_, sFilters, err := a.StorageFiltersForList(sub, policyTY.ResourceSource)
	if err != nil {
		t.Fatal(err)
	}
	sources := []interface{}{
		&sourceTY.Source{ID: "a", GatewayID: "mysensor", NodeID: "1", SourceID: "s"},
		&sourceTY.Source{ID: "b", GatewayID: "mysensor", NodeID: "2", SourceID: "s"},
	}
	sm := filterUtils.Filter(sources, sFilters, false)
	if len(sm) != 1 || sm[0].(*sourceTY.Source).NodeID != "2" {
		t.Fatalf("source filter wrong: matched=%d filters=%+v", len(sm), sFilters)
	}
}

func TestStorageFiltersForList_ReadwritePlusDeny(t *testing.T) {
	p := policyTY.Policy{
		ID: "p1",
		Statements: []policyTY.Statement{
			{
				Effect:    policyTY.EffectAllow,
				Actions:   []string{"*"},
				Resources: []string{"gateway", "node", "source", "field", "metric"},
			},
			{
				Effect:    policyTY.EffectDeny,
				Actions:   []string{"*"},
				Resources: []string{"node:mysensor.1"},
			},
		},
	}
	a := mockAPIWithPolicy(t, "u1", p)
	sub := Subject{UserID: "u1"}

	_, filters, err := a.StorageFiltersForList(sub, policyTY.ResourceField)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("deny-only filters: %+v", filters)

	entities := []interface{}{
		&fieldTY.Field{ID: "a", GatewayID: "mysensor", NodeID: "1", SourceID: "s", FieldID: "t"},
		&fieldTY.Field{ID: "b", GatewayID: "mysensor", NodeID: "2", SourceID: "s", FieldID: "t"},
		&fieldTY.Field{ID: "c", GatewayID: "other", NodeID: "9", SourceID: "s", FieldID: "t"},
	}
	matched := filterUtils.Filter(entities, filters, false)
	if len(matched) != 2 {
		t.Fatalf("want 2 fields (all except node1), got %d; filters=%+v", len(matched), filters)
	}
}

func TestAllowed_GetFieldUnderSibling(t *testing.T) {
	p := policyTY.Policy{
		ID: "p1",
		Statements: []policyTY.Statement{
			{
				Effect:  policyTY.EffectAllow,
				Actions: []string{"*"},
				Resources: []string{
					"field:mysensor.*",
					"source:mysensor.*",
					"node:mysensor.*",
				},
			},
			{
				Effect:    policyTY.EffectDeny,
				Actions:   []string{"*"},
				Resources: []string{"node:mysensor.1"},
			},
		},
	}
	a := mockAPIWithPolicy(t, "u1", p)
	sub := Subject{UserID: "u1"}

	if err := a.Allowed(sub, "get", "field:mysensor.2.s.t"); err != nil {
		t.Fatalf("sibling field get: %v", err)
	}
	if err := a.Allowed(sub, "get", "field:mysensor.1.s.t"); err == nil {
		t.Fatal("denied node field should fail")
	}
	if err := a.Allowed(sub, "list", "field"); err != nil {
		t.Fatalf("list field: %v", err)
	}
	if err := a.Allowed(sub, "get", "source:mysensor.2.s"); err != nil {
		t.Fatalf("sibling source get: %v", err)
	}
	if err := a.Allowed(sub, "get", "source:mysensor.1.s"); err == nil {
		t.Fatal("denied node source should fail")
	}
}

// Exact policy from production memory_db (test user): only gateway + node allow.
func TestLiveTestPolicy_NodeAllowCascadesToSourceField(t *testing.T) {
	p := policyTY.Policy{
		ID: "test",
		Statements: []policyTY.Statement{
			{
				Effect:  policyTY.EffectAllow,
				Actions: []string{"get", "list"},
				Resources: []string{
					"gateway:mysensor",
					"node:mysensor.*",
				},
			},
			{
				Effect:    policyTY.EffectDeny,
				Actions:   []string{"get", "list"},
				Resources: []string{"node:mysensor.1"},
			},
		},
	}
	a := mockAPIWithPolicy(t, "u1", p)
	sub := Subject{UserID: "u1"}

	// list APIs for child kinds must be allowed (then filtered)
	for _, kind := range []string{"source", "field", "node", "gateway"} {
		if err := a.Allowed(sub, "list", kind); err != nil {
			t.Fatalf("list %s: %v", kind, err)
		}
	}

	// filters: allow under mysensor, exclude node 1
	_, patterns, err := a.ResourceNamesForList(sub, "field")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("field patterns: %v", patterns)
	allow, deny := splitAllowDenyPatterns(patterns)
	if len(allow) == 0 && !containsPattern(patterns, "mysensor") {
		// either explicit allow mysensor.* or hasWildcard with deny only
		if len(deny) == 0 {
			t.Fatalf("expected allow and/or deny patterns, got %v", patterns)
		}
	}
	foundDeny := false
	for _, d := range deny {
		if strings.TrimPrefix(d, "!") == "mysensor.1" {
			foundDeny = true
		}
	}
	if !foundDeny {
		t.Fatalf("expected !mysensor.1 deny, got %v", patterns)
	}

	_, filters, err := a.StorageFiltersForList(sub, "field")
	if err != nil {
		t.Fatal(err)
	}
	entities := []interface{}{
		&fieldTY.Field{ID: "a", GatewayID: "mysensor", NodeID: "1", SourceID: "s", FieldID: "t"},
		&fieldTY.Field{ID: "b", GatewayID: "mysensor", NodeID: "2", SourceID: "s", FieldID: "t"},
	}
	matched := filterUtils.Filter(entities, filters, false)
	if len(matched) != 1 || matched[0].(*fieldTY.Field).NodeID != "2" {
		t.Fatalf("want only node2 field, got %d filters=%+v", len(matched), filters)
	}

	if err := a.Allowed(sub, "get", "field:mysensor.2.s.t"); err != nil {
		t.Fatalf("get sibling field: %v", err)
	}
	if err := a.Allowed(sub, "get", "field:mysensor.1.s.t"); err == nil {
		t.Fatal("get denied node field should fail")
	}
}

func containsPattern(patterns []string, sub string) bool {
	for _, p := range patterns {
		if strings.Contains(p, sub) {
			return true
		}
	}
	return false
}
