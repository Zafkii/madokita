package icons

import "embed"

// FS contains every SVG stored in this directory. New icons can be dropped in
// and picked up by the editor after a rebuild.
//
//go:embed *.svg
var FS embed.FS
