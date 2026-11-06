package cmd

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/LarsArtmann/art-dupl/baseline"
	"github.com/LarsArtmann/art-dupl/cache"
	"github.com/LarsArtmann/art-dupl/config"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
	"github.com/zeebo/xxh3"
)

// baselineVersionMustMatchFlag refuses a `check` run whose analyzer identity
// differs from the one recorded in the baseline (issue #13).
const baselineVersionMustMatchFlag = "baseline-version-must-match"

// analyzerPolicyParts renders the detection settings that change the set or
// the hashes of the reported groups. Anything that alters hashing or which
// detector produces the group belongs here; settings that only affect
// presentation (sorting, actionability, thresholds for other filters) do not.
func analyzerPolicyParts(cfg *config.Config) string {
	return fmt.Sprintf(
		"cache=%d;mode=%s;typeAware=%t;suggestGenerics=%t;maxChildren=%d;methods=%s",
		cache.CacheVersion,
		cfg.DetectionMode.String(),
		cfg.TypeAware,
		cfg.SuggestGenerics,
		cfg.MaxChildrenSerial,
		detectionMethodsPolicyKey(cfg.DetectionMethods),
	)
}

// detectionMethodsPolicyKey renders the detection-method set order-
// independently: `-m hash,art-dupl` and `-m art-dupl,hash` run the same
// detectors, but a different detector SET produces different group hashes
// (hash-detected groups would read as new against an art-dupl-only baseline).
func detectionMethodsPolicyKey(methods []config.DetectionMethod) string {
	names := make([]string, 0, len(methods))
	for _, method := range methods {
		names = append(names, method.String())
	}

	sort.Strings(names)

	return strings.Join(names, ",")
}

// analyzerIdentity describes what produced a run's clone-group hashes: the
// algorithm revision (cache version) plus the hash-affecting settings, folded
// into one policy fingerprint. The binary version is recorded alongside it as
// provenance but does not participate in compatibility.
func analyzerIdentity(cfg *config.Config) baseline.AnalyzerIdentity {
	return baseline.AnalyzerIdentity{
		Version:      GetVersion(),
		CacheVersion: cache.CacheVersion,
		Policy:       fmt.Sprintf("%016x", xxh3.HashString(analyzerPolicyParts(cfg))),
	}
}

// warnIncomparableBaseline prints a stderr notice when the running analyzer
// cannot be verified against the baseline's recorded identity, without
// changing the exit code: the strict gate stays opt-in
// (--baseline-version-must-match). Silence was the failure mode issue #13
// reported — a whole tree of "new" clones and no statement about why.
func warnIncomparableBaseline(
	w io.Writer,
	bf *baseline.File,
	cfg *config.Config,
	path string,
) {
	running := analyzerIdentity(cfg)

	switch {
	case bf.Analyzer == nil:
		fmt.Fprintf(
			w,
			"warning: baseline %s carries no analyzer identity (recorded before identity tracking); "+
				"its hashes cannot be verified against %s. Re-record it with `art-dupl baseline`.\n",
			path, running.String(),
		)
	case !bf.Analyzer.CompatibleWith(running):
		fmt.Fprintf(
			w,
			"warning: baseline %s was recorded by %s; running %s: clone-group hashes are not comparable, "+
				"so every recorded group reads as new. Re-record the baseline with the running binary, "+
				"or pass --%s to make this an error.\n",
			path, bf.Analyzer.String(), running.String(), baselineVersionMustMatchFlag,
		)
	}
}

// ensureBaselineIdentityMatches refuses to compare against a baseline whose
// hashing policy differs from the running analyzer's. Group hashes from two
// different algorithm revisions or detection settings live in different hash
// spaces: every recorded group looks new, and the CI gate reports a whole-tree
// regression with no code change. A baseline that predates identity recording
// cannot be verified and is refused as well — the caller asked for a
// guarantee, not a best guess.
func ensureBaselineIdentityMatches(
	bf *baseline.File,
	cfg *config.Config,
	path string,
) error {
	running := analyzerIdentity(cfg)

	if bf.Analyzer == nil {
		return duplerrors.NewValidationError(
			fmt.Sprintf(
				"baseline %s carries no analyzer identity (recorded before identity tracking); "+
					"re-record it with `art-dupl baseline`, or drop --%s to compare unverified",
				path, baselineVersionMustMatchFlag,
			), nil,
		)
	}

	if bf.Analyzer.CompatibleWith(running) {
		return nil
	}

	return duplerrors.NewValidationError(
		fmt.Sprintf(
			"baseline %s was recorded by %s; running %s: clone-group hashes are not comparable, "+
				"re-record the baseline with the running binary or drop --%s",
			path, bf.Analyzer.String(), running.String(), baselineVersionMustMatchFlag,
		), nil,
	)
}
