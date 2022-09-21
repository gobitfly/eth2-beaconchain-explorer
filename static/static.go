package static

import "embed"

var (
	//go:embed *
	Files embed.FS
)
