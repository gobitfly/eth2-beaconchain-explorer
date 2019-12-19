package version

import (
	"runtime"
)

// Build information. Populated at build-time
var (
	Version   = "undefined"
	GitDate   = "undefined"
	GitCommit = "undefined"
	BuildDate = "undefined"
	GoVersion = runtime.Version()
)
