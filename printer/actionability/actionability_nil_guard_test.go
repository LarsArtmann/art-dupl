package actionability

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// Real-source tests for the nil-guard-after-helper pattern (plan T13). All
// fixtures parse through golang.ParseSnippet so the node shapes mirror
// production serialization exactly — the session-3 lesson: hand-built trees
// lie in both directions (they masked the append-collector-tail child-order
// defect for an entire session).
func TestNilGuardAfterHelperRealSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  string
		want bool
	}{
		{
			name: "mined branching-flow prelude matches (assign + nil guard)",
			src: `indexIdent := astutil.IndexIdent(index)
if indexIdent == nil {
	return false
}`,
			want: true,
		},
		{
			name: "nil-first comparison order matches",
			src: `result := lookup(key)
if nil == result {
	return nil
}`,
			want: true,
		},
		{
			name: "not-nil guard with bare return matches",
			src: `fn := resolve(name)
if fn != nil {
	return
}`,
			want: true,
		},
		{
			name: "guard with trailing bare return matches",
			src: `v := findNode(root, path)
if v == nil {
	return false
}
return true`,
			want: true,
		},
		{
			name: "two guards and one assignment matches",
			src: `index := at(index)
name := exprName(x)
if name == nil {
	return ""
}
if index == nil {
	return ""
}`,
			want: true,
		},
		{
			name: "guard body closes a resource: NOT matched (FP guard)",
			src: `tx := begin()
if tx != nil {
	return tx.Rollback()
}
return nil`,
			want: false,
		},
		{
			name: "guard body constructs an error: NOT matched",
			src: `conn := dial(addr)
if conn == nil {
	return fmt.Errorf("dial %s", addr)
}`,
			want: false,
		},
		{
			name: "guard body mutates state: NOT matched",
			src: `h := acquire()
if h == nil {
	s.failed = true
	return false
}`,
			want: false,
		},
		{
			name: "assignment carries a closure: NOT matched",
			src: `p := transform(data, func(n int) int { return n + 1 })
if p == nil {
	return false
}`,
			want: false,
		},
		{
			name: "non-nil comparison guard: NOT matched",
			src: `n := count(items)
if n == 0 {
	return false
}`,
			want: false,
		},
		{
			name: "else branch present: NOT matched",
			src: `x := find(key)
if x == nil {
	return false
} else {
	return true
}`,
			want: false,
		},
		{
			name: "lone guard without assignment: NOT matched (guard-clause territory)",
			src: `if params == nil {
	return false
}`,
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			nodes, err := golang.ParseSnippet(
				golang.SnippetStatements, tc.src, "fixture.go", 0, golang.DetectionModeSemantic,
			)
			if err != nil {
				t.Fatalf("ParseSnippet error: %v", err)
			}

			if len(nodes) == 0 {
				t.Fatal("ParseSnippet returned no nodes")
			}

			seq := make([]*domain.CloneNode, len(nodes))
			for i, n := range nodes {
				seq[i] = cloneNodeFrom(n)
			}

			label, action := evaluateActionabilityWithDisabled([][]*domain.CloneNode{seq}, disabledExcept(
				PatternNilGuardAfterHelper,
			))
			if tc.want && action != domain.NonActionable {
				t.Errorf("expected NonActionable, got %v (label %q)", action, label)
			}

			if !tc.want && label == PatternNilGuardAfterHelper {
				t.Errorf("pattern matched but should not have")
			}
		})
	}
}

// TestOptionApplyLoopSeededRealSource pins the option-apply v2 extension
// against the four gap families measured in the survivor labeling
// (2384b605, 87e6d65d, 754d60f8, 67761296) plus negative controls.
func TestOptionApplyLoopSeededRealSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  string
		want bool
	}{
		{
			name: "constructor call seeding matches (gap 2384b605)",
			src: `cfg := defaultConfig()
for _, opt := range opts {
	opt(&cfg)
}`,
			want: true,
		},
		{
			name: "composite literal seeding matches (gap 87e6d65d)",
			src: `cfg := engineConfig{syncWrites: true}
for _, opt := range engineOpts {
	opt(&cfg)
}`,
			want: true,
		},
		{
			name: "seeding with blank line matches (gap 754d60f8)",
			src: `cfg := defaultConfig()

for _, opt := range opts {
	opt(&cfg)
}

return cfg`,
			want: true,
		},
		{
			name: "lookup guard before loop matches (gap 67761296)",
			src: `if !ok {
	return
}

for _, opt := range opts {
	opt(d)
}`,
			want: true,
		},
		{
			name: "error guard before loop matches",
			src: `if err != nil {
	return err
}

for _, opt := range opts {
	opt(&cfg)
}
return cfg`,
			want: true,
		},
		{
			name: "seeding with closure argument does NOT match",
			src: `cfg := newConfig(func(c *Config) { c.retries = 3 })
for _, opt := range opts {
	opt(&cfg)
}`,
			want: false,
		},
		{
			name: "guard with non-trivial body does NOT match",
			src: `if !ok {
	s.reset()
	return
}

for _, opt := range opts {
	opt(d)
}`,
			want: false,
		},
		{
			name: "defaulting logic after the loop does NOT match",
			src: `cfg := defaultConfig()
for _, opt := range opts {
	opt(&cfg)
}
if cfg.Name == "" {
	cfg.Name = "default"
}
return cfg`,
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			nodes, err := golang.ParseSnippet(
				golang.SnippetStatements, tc.src, "fixture.go", 0, golang.DetectionModeSemantic,
			)
			if err != nil {
				t.Fatalf("ParseSnippet error: %v", err)
			}

			if len(nodes) == 0 {
				t.Fatal("ParseSnippet returned no nodes")
			}

			seq := make([]*domain.CloneNode, len(nodes))
			for i, n := range nodes {
				seq[i] = cloneNodeFrom(n)
			}

			label, action := evaluateActionabilityWithDisabled([][]*domain.CloneNode{seq}, disabledExcept(
				PatternOptionApplyLoop,
			))
			if tc.want && action != domain.NonActionable {
				t.Errorf("expected NonActionable, got %v (label %q)", action, label)
			}

			if !tc.want && label == PatternOptionApplyLoop {
				t.Errorf("pattern matched but should not have")
			}
		})
	}
}
