//go:build web
// +build web

package web

import (
	assets "github.com/mycontroller-org/server/v2/pkg/http_router/web-console/actual"
)

// StaticFiles provides http filesystem with static files for UI
var StaticFiles = assets.FS(false)
