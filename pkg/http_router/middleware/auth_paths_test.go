package handler

import (
	"testing"

	handlerTY "github.com/mycontroller-org/server/v2/pkg/types/web_handler"
)

func TestIsNonRestricted(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"/api/status", true},
		{"/api/status/", true},
		{"/api/user/login", true},
		{"/api/oauth/token", true},
		// prefix look-alikes must stay authenticated: /api/user/{id} would match
		// "/api/user/registration" as a prefix
		{"/api/user/registrationX", false},
		{"/api/user/logins", false},
		{"/api/statuses", false},
		{"/api/user/profile", false},
		{"/api/gateway", false},
		{"/api/user/some-user-id", false},
		// share directories are path trees
		{handlerTY.InsecureShareDirWebHandlerPath + "/some/file.txt", true},
		{handlerTY.SecureShareDirWebHandlerPath + "/some/file.txt", false},
	}

	for _, c := range cases {
		if got := isNonRestricted(c.path); got != c.want {
			t.Errorf("isNonRestricted(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestIsOwnProfilePath(t *testing.T) {
	cases := map[string]bool{
		"/api/user/profile":   true,
		"/api/user/profile/":  true,
		"/api/user/profileX":  false,
		"/api/user/profile/x": false,
		"/api/user/other-id":  false,
	}
	for path, want := range cases {
		if got := isOwnProfilePath(path); got != want {
			t.Errorf("isOwnProfilePath(%q) = %v, want %v", path, got, want)
		}
	}
}
