//go:build !embedui

package web

import _ "embed"

// Placeholder archive used when the UI was not packed into the binary.
// Release builds use -tags=embedui and assets_packed.go instead.
//
//go:embed assets/placeholder.zip
var embeddedZIP []byte
