package css

import "embed"

//go:embed *.css **/*.css
var Files embed.FS
