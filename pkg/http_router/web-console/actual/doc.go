// Package assets holds embedded web-console static files.
//
// When building with -tags=web, generate the real implementation first:
//
//	esc -pkg assets -o pkg/http_router/web-console/actual/generated_assets.go \
//	    -prefix web-console/build web-console/build
//
// See scripts/build_web_console.sh. generated_assets.go is gitignored.
package assets
