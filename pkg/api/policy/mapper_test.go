package policy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
)

func TestMapRequestUserProfileExact(t *testing.T) {
	profile := MapRequest(httptest.NewRequest(http.MethodGet, "/api/user/profile", nil))
	if profile.Name != "" || profile.Kind != policyTY.ResourceUser || profile.Action != policyTY.ActionGet {
		t.Fatalf("profile: %+v", profile)
	}

	named := MapRequest(httptest.NewRequest(http.MethodGet, "/api/user/profileX", nil))
	if named.Name != "profileX" || named.Action != policyTY.ActionGet {
		t.Fatalf("named profileX should be a user get, got %+v", named)
	}
}
