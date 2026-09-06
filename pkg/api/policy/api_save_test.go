package policy

import (
	"errors"
	"testing"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	svcTokenTY "github.com/mycontroller-org/server/v2/pkg/types/service_token"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

type policyMemStore struct {
	policies map[string]policyTY.Policy
}

func newPolicyMemStore() *policyMemStore {
	return &policyMemStore{policies: map[string]policyTY.Policy{}}
}

func (s *policyMemStore) Name() string { return "test" }
func (s *policyMemStore) Ping() error  { return nil }
func (s *policyMemStore) Close() error { return nil }
func (s *policyMemStore) Insert(string, interface{}) error {
	return errors.New("not implemented")
}
func (s *policyMemStore) Update(string, interface{}, []storageTY.Filter) error {
	return errors.New("not implemented")
}
func (s *policyMemStore) Find(string, interface{}, []storageTY.Filter, *storageTY.Pagination) (*storageTY.Result, error) {
	return nil, errors.New("not implemented")
}
func (s *policyMemStore) Delete(string, []storageTY.Filter) (int64, error) {
	return 0, errors.New("not implemented")
}
func (s *policyMemStore) Pause() error                                 { return nil }
func (s *policyMemStore) Resume() error                                { return nil }
func (s *policyMemStore) ClearDatabase() error                         { return nil }
func (s *policyMemStore) DoStartupImport() (bool, string, string)      { return false, "", "" }

func (s *policyMemStore) FindOne(entityName string, out interface{}, filters []storageTY.Filter) error {
	if entityName != types.EntityPolicy || len(filters) == 0 {
		return storageTY.ErrNoDocuments
	}
	id, _ := filters[0].Value.(string)
	p, ok := s.policies[id]
	if !ok {
		return storageTY.ErrNoDocuments
	}
	*(out.(*policyTY.Policy)) = p
	return nil
}

func (s *policyMemStore) Upsert(entityName string, data interface{}, _ []storageTY.Filter) error {
	p, ok := data.(*policyTY.Policy)
	if !ok || entityName != types.EntityPolicy {
		return errors.New("invalid upsert")
	}
	s.policies[p.ID] = *p
	return nil
}

func testPolicyAPI(store *policyMemStore) *API {
	c := newCache()
	c.setLoaders(
		func(id string) (*userTY.User, error) { return nil, ErrUserNotFound },
		func(id string) (*policyTY.Policy, error) { return nil, ErrUserNotFound },
		func(id string) (*svcTokenTY.ServiceToken, error) { return nil, ErrTokenNotFound },
		func() ([]policyTY.Policy, error) { return nil, nil },
	)
	return &API{storage: store, cache: c}
}

func TestSaveRejectsSystemPolicyEdit(t *testing.T) {
	store := newPolicyMemStore()
	store.policies["admin"] = policyTY.Policy{ID: "admin", System: true}
	a := testPolicyAPI(store)

	err := a.Save(&policyTY.Policy{ID: "admin", Description: "hacked", System: false})
	if !errors.Is(err, ErrSystemPolicyImmutable) {
		t.Fatalf("Save system policy: got %v, want %v", err, ErrSystemPolicyImmutable)
	}
}

func TestSaveRejectsBuiltInID(t *testing.T) {
	a := testPolicyAPI(newPolicyMemStore())
	err := a.Save(&policyTY.Policy{ID: policyTY.PolicyReadOnly, Description: "custom readonly"})
	if !errors.Is(err, ErrSystemPolicyImmutable) {
		t.Fatalf("Save built-in id: got %v, want %v", err, ErrSystemPolicyImmutable)
	}
}

func TestSaveRejectsSystemFlagOnCustomPolicy(t *testing.T) {
	a := testPolicyAPI(newPolicyMemStore())
	err := a.Save(&policyTY.Policy{ID: "living-room", System: true})
	if !errors.Is(err, ErrSystemFlagNotAllowed) {
		t.Fatalf("Save custom with system: got %v, want %v", err, ErrSystemFlagNotAllowed)
	}
}

func TestSavePersistsCustomPolicyWithoutSystem(t *testing.T) {
	store := newPolicyMemStore()
	a := testPolicyAPI(store)
	if err := a.Save(&policyTY.Policy{ID: "living-room", Description: "ok"}); err != nil {
		t.Fatalf("Save custom: %v", err)
	}
	got := store.policies["living-room"]
	if got.System {
		t.Fatal("expected system=false on stored custom policy")
	}
	if got.Description != "ok" {
		t.Fatalf("description=%q", got.Description)
	}
}

func TestImportSkipsBuiltInAndStripsSystemFlag(t *testing.T) {
	store := newPolicyMemStore()
	store.policies["admin"] = policyTY.Policy{ID: "admin", System: true, Description: "original"}
	a := testPolicyAPI(store)

	if err := a.Import(policyTY.Policy{ID: "admin", Description: "from backup", System: true}); err != nil {
		t.Fatalf("Import built-in: %v", err)
	}
	if store.policies["admin"].Description != "original" {
		t.Fatalf("built-in was overwritten: %+v", store.policies["admin"])
	}

	if err := a.Import(policyTY.Policy{ID: "custom-1", System: true, Description: "restored"}); err != nil {
		t.Fatalf("Import custom: %v", err)
	}
	got := store.policies["custom-1"]
	if got.System {
		t.Fatal("import left system=true on custom policy")
	}
	if got.Description != "restored" {
		t.Fatalf("description=%q", got.Description)
	}
}
