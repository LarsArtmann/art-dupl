package cmd

import (
	"testing"
)

// Issue #8: dev binaries carry no provenance. resolveBuildInfo merges the
// ldflag values with debug.ReadBuildInfo fallbacks so `go install` and
// checkout builds self-identify.
func TestResolveBuildInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string

		ldVersion   string
		ldCommit    string
		ldDate      string
		moduleVer   string
		vcsRevision string
		vcsTime     string

		wantVersion string
		wantCommit  string
		wantDate    string
	}{
		{
			name:        "release ldflags win",
			ldVersion:   "0.9.0",
			ldCommit:    "8ce4b2d1234567890",
			ldDate:      "2026-10-01",
			moduleVer:   "(devel)",
			vcsRevision: "fffffffffffffff",
			wantVersion: "0.9.0",
			wantCommit:  "8ce4b2d1234567890",
			wantDate:    "2026-10-01",
		},
		{
			name:        "go install module version replaces dev",
			ldVersion:   "dev",
			moduleVer:   "v0.9.0",
			wantVersion: "v0.9.0",
			wantCommit:  "unknown",
			wantDate:    "unknown",
		},
		{
			name:        "checkout build falls back to vcs revision",
			ldVersion:   "dev",
			moduleVer:   "(devel)",
			vcsRevision: "8ce4b2d1234567890",
			vcsTime:     "2026-10-15T16:00:00Z",
			wantVersion: "dev-8ce4b2d",
			wantCommit:  "8ce4b2d1234567890",
			wantDate:    "2026-10-15T16:00:00Z",
		},
		{
			name:        "module version preferred over vcs revision",
			ldVersion:   "dev",
			moduleVer:   "v0.9.1",
			vcsRevision: "8ce4b2d1234567890",
			wantVersion: "v0.9.1",
			wantCommit:  "8ce4b2d1234567890",
			wantDate:    "unknown",
		},
		{
			name:        "no buildinfo at all keeps dev",
			ldVersion:   "dev",
			wantVersion: "dev",
			wantCommit:  "unknown",
			wantDate:    "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := resolveBuildInfo(tt.ldVersion, tt.ldCommit, tt.ldDate, tt.moduleVer, tt.vcsRevision, tt.vcsTime)

			if got.version != tt.wantVersion {
				t.Errorf("version = %q, want %q", got.version, tt.wantVersion)
			}

			if got.commit != tt.wantCommit {
				t.Errorf("commit = %q, want %q", got.commit, tt.wantCommit)
			}

			if got.date != tt.wantDate {
				t.Errorf("date = %q, want %q", got.date, tt.wantDate)
			}
		})
	}
}

// GetVersion appends the short commit suffix for release builds (matching
// the nix release stamp shape `0.9.0-8ce4b2d`) but never duplicates the
// revision already embedded in the `dev-<rev>` buildinfo fallback.
func TestVersionCommitSuffix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string

		version string
		commit  string

		want string
	}{
		{"release build appends short commit", "0.9.0", "8ce4b2d1234567890", "0.9.0-8ce4b2d"},
		{"dev-rev fallback does not duplicate", "dev-8ce4b2d", "8ce4b2d1234567890", "dev-8ce4b2d"},
		{"unknown commit stays bare", "dev", "unknown", "dev"},
		{"short commit stays bare", "dev", "abc", "dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			version := tt.version

			if suffix := commitSuffix(tt.version, tt.commit); suffix != "" {
				version += "-" + suffix
			}

			if version != tt.want {
				t.Errorf("version = %q, want %q", version, tt.want)
			}
		})
	}
}

func TestShortRevision(t *testing.T) {
	t.Parallel()

	tests := []struct {
		rev  string
		want string
	}{
		{"8ce4b2d1234567890", "8ce4b2d"},
		{"8ce4b2d", "8ce4b2d"},
		{"", ""},
	}

	for _, tt := range tests {
		if got := shortRevision(tt.rev); got != tt.want {
			t.Errorf("shortRevision(%q) = %q, want %q", tt.rev, got, tt.want)
		}
	}
}
