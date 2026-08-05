package cmd

import (
	"fmt"
	"io"
	"runtime"
)

// Version information - can be overridden during build.
var (
	Version = "dev"     //nolint:gochecknoglobals // Build-time variables meant to be overridden
	Commit  = "unknown" //nolint:gochecknoglobals
	Date    = "unknown" //nolint:gochecknoglobals
)

// GetVersion returns version information including build info.
func GetVersion() string {
	version := Version

	if Commit != "unknown" && len(Commit) > 7 {
		version += "-" + Commit[:7]
	}

	return version
}

// GetCommit returns commit hash.
func GetCommit() string {
	return Commit
}

// GetBuildDate returns build date.
func GetBuildDate() string {
	return Date
}

// PrintVersion prints version information to w.
func PrintVersion(w io.Writer) {
	fmt.Fprintf(w,
		"art-dupl version %s\n",
		GetVersion(),
	)
	fmt.Fprintf(w,
		"Built with %s %s/%s\n",
		runtime.Compiler,
		runtime.GOOS,
		runtime.GOARCH,
	)
}
