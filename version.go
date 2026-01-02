package main

import (
	"fmt"
	"runtime"
)

// Version information - can be overridden during build.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

// GetVersion returns version information including build info.
func GetVersion() string {
	version := Version

	if Commit != "unknown" && len(Commit) > 7 {
		version += "-" + Commit[:7]
	}

	return version
}

// GetCommit returns the commit hash.
func GetCommit() string {
	return Commit
}

// GetBuildDate returns the build date.
func GetBuildDate() string {
	return Date
}

// PrintVersion prints version information.
func PrintVersion() {
	fmt.Printf("art-dupl version %s\n", GetVersion())
	fmt.Printf("Built with %s %s/%s\n", runtime.Compiler, runtime.GOOS, runtime.GOARCH)
}
