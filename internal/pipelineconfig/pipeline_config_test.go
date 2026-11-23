package pipelineconfig

import (
	"os"
	"slices"
	"strings"
	"testing"
)

// buildflowConfigPath is the project-override file the BuildFlow daemon and
// every pipeline load. It has been deleted twice by an unidentified destroyer
// (jsonv2 recurrences #20 and #21, commits 8ce90a47 and 57782482): each
// deletion silently disabled the skip_steps overrides, so the very next
// `buildflow --fix` re-poisoned the six v1-API JSON files and flipped the
// go-line pin. This gate turns that deletion into a red suite instead of
// silent config loss.
const buildflowConfigPath = "../../.buildflow.yml"

// restoreRitual is the verified remediation for a missing or stripped file.
const restoreRitual = "git restore --source=c00f3222 --worktree --staged .buildflow.yml"

func readBuildflowConfig(t *testing.T) string {
	t.Helper()

	data, err := os.ReadFile(buildflowConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf(".buildflow.yml is MISSING (destroyer struck again: recurrences #20/#21). Restore it: %s. If its removal is intentional, update this gate to the new contract.", restoreRitual)
		}
		t.Fatalf("read .buildflow.yml: %v", err)
	}

	return string(data)
}

// skipStepsBlock returns the list items directly under the top-level
// skip_steps key (lines like "  - name"), stopping at the next top-level key.
func skipStepsBlock(t *testing.T, content string) []string {
	t.Helper()

	var items []string
	inBlock := false

	for line := range strings.SplitSeq(content, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "skip_steps:"):
			inBlock = true
		case inBlock && (trimmed == "" || strings.HasPrefix(trimmed, "#")):
			continue
		case inBlock && strings.HasPrefix(trimmed, "- "):
			items = append(items, strings.TrimPrefix(trimmed, "- "))
		case inBlock:
			return items
		}
	}

	return items
}

func TestBuildflowConfigExists(t *testing.T) {
	readBuildflowConfig(t)
}

func TestBuildflowConfigKeepsGoAutoUpgradeSkipped(t *testing.T) {
	items := skipStepsBlock(t, readBuildflowConfig(t))

	for _, required := range []string{"go-auto-upgrade", "golangci-lint-auto-configure"} {
		found := slices.Contains(items, required)
		if !found {
			t.Fatalf("skip_steps lost %q (items: %v). Without the skip, buildflow --fix re-poisons the v1-API JSON files (recurrences #17-#21). Restore: %s. If the skip removal is intentional (owner Q1 resolved), update this gate.", required, items, restoreRitual)
		}
	}
}
