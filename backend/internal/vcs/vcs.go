package vcs

import (
	"runtime/debug"
)

// Version returns the main module version from embedded build information, which
// may be "(devel)" for local builds. It returns an empty string if build information
// is unavailable.
func Version() string {
	bi, ok := debug.ReadBuildInfo()
	if ok {
		return bi.Main.Version
	}

	return ""
}
