package frontenddist

import "embed"

//go:embed all:dist
var FileSystem embed.FS
