package desktop

import "embed"

//go:embed all:frontend/dist
var assets embed.FS
