package cmd

import (
	"fmt"
	"io"
	"runtime"
	"runtime/debug"
	"sync"
)

// Version information - can be overridden during build.
var (
	Version = "dev"     //nolint:gochecknoglobals // Build-time variables meant to be overridden
	Commit  = "unknown" //nolint:gochecknoglobals
	Date    = "unknown" //nolint:gochecknoglobals
)

// buildInfo carries the ldflags values merged with debug.ReadBuildInfo
// fallbacks so every binary self-identifies, even without release ldflags
// (issue #8: a `dev` binary's reports were unactionable because the build
// could not be reconstructed).
type buildInfo struct {
	version string
	commit  string
	date    string
}

// resolveBuildInfo merges ldflag values with buildinfo fallbacks:
//
//   - version: when the ldflag override did not happen ("dev"), a
//     `go install` module-cache build reports its module version (e.g.
//     "v0.9.0"); a checkout build reports "dev-<short vcs revision>".
//   - commit/date: when unknown, fall back to the embedded vcs.revision /
//     vcs.time build settings (present unless -buildvcs=false).
//
// Pure in its inputs so the fallback logic is unit-testable; readBuildInfo
// is the thin production wrapper.
func resolveBuildInfo(
	ldVersion, ldCommit, ldDate, moduleVersion, vcsRevision, vcsTime string,
) buildInfo {
	info := buildInfo{version: ldVersion, commit: ldCommit, date: ldDate}

	if info.version == "dev" && moduleVersion != "" && moduleVersion != "(devel)" {
		info.version = moduleVersion
	}

	if info.version == "dev" && vcsRevision != "" {
		info.version = "dev-" + shortRevision(vcsRevision)
	}

	if info.commit == "unknown" && vcsRevision != "" {
		info.commit = vcsRevision
	}

	if info.date == "unknown" && vcsTime != "" {
		info.date = vcsTime
	}

	return info
}

// shortRevision truncates a full VCS revision to the conventional 7-char
// short form (hex git hashes are unambiguous at that length).
func shortRevision(rev string) string {
	if len(rev) > 7 {
		return rev[:7]
	}

	return rev
}

var effectiveBuildInfo = sync.OnceValue(func() buildInfo { //nolint:gochecknoglobals // computed once from immutable build metadata
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return buildInfo{version: Version, commit: Commit, date: Date}
	}

	var vcsRevision, vcsTime string

	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			vcsRevision = setting.Value
		case "vcs.time":
			vcsTime = setting.Value
		}
	}

	return resolveBuildInfo(Version, Commit, Date, info.Main.Version, vcsRevision, vcsTime)
})

// GetVersion returns version information including build info.
func GetVersion() string {
	info := effectiveBuildInfo()

	version := info.version
	if info.commit != "unknown" && len(info.commit) > 7 {
		version += "-" + shortRevision(info.commit)
	}

	return version
}

// GetCommit returns commit hash.
func GetCommit() string {
	return effectiveBuildInfo().commit
}

// GetBuildDate returns build date.
func GetBuildDate() string {
	return effectiveBuildInfo().date
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
