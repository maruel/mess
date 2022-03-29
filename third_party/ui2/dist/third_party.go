package dist

import "embed"

//go:embed *.css *.js *.html
var FS embed.FS
