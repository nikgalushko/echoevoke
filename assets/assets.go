package assets

import "embed"

//go:embed sql
var SQL embed.FS

//go:embed html
var HTML embed.FS
