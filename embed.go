package williamfindlaycom

import "embed"

//go:embed all:static all:templates
var Embedded embed.FS
