//go:build embedui

package web

import _ "embed"

// Packed production UI. Created by scripts/cmd/pack_web_console before compile.
//
//go:embed assets/web_console.zip
var embeddedZIP []byte
