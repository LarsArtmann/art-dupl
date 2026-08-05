package actionability

import "testing"

func TestIsFormatSpecifierDifference(t *testing.T) {
	for _, tc := range []struct {
		name string
		a, b string
		want bool
	}{
		// Same specifiers — no difference
		{a: "%d", b: "%d", want: false},
		{a: "error: %s", b: "warning: %s", want: false},
		{a: "%d items in %s", b: "%d bytes in %s", want: false},

		// Different verbs
		{a: "%d", b: "%s", want: true},
		{a: "got %d", b: "got %v", want: true},

		// Different argument count (extra specifier)
		{a: "%d", b: "%d %s", want: true},

		// Width specifier difference
		{a: "%5d", b: "%3d", want: true},
		{a: "%10s", b: "%5s", want: true},

		// Precision difference
		{a: "%.2f", b: "%.3f", want: true},

		// Flags difference
		{a: "%+d", b: "%d", want: true},

		// Argument index difference
		{a: "%[1]d", b: "%[2]d", want: true},
		{a: "%[1]s %s", b: "%s %[1]s", want: true},

		// Literal percent (%%) is NOT a specifier
		{a: "100%%d", b: "100%%s", want: false},
		{a: "progress: 50%%", b: "progress: 75%%", want: false},

		// %% followed by a real verb
		{a: "100%% done %d", b: "100%% done %s", want: true},

		// No specifiers at all
		{a: "hello", b: "world", want: false},
		{a: "", b: "", want: false},

		// Mixed literal and real specifiers
		{a: "%d%% complete", b: "%s%% complete", want: true},
		{a: "%d%% complete", b: "%d%% done", want: false},

		// Different number of specifiers
		{a: "%d %s", b: "%d", want: true},
	} {
		t.Run(tc.a+"_vs_"+tc.b, func(t *testing.T) {
			got := isFormatSpecifierDifference(tc.a, tc.b)
			if got != tc.want {
				t.Fatalf("isFormatSpecifierDifference(%q, %q) = %v, want %v",
					tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestExtractFormatSpecifiers(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  []string
	}{
		{"%d", []string{"%d"}},
		{"no verbs", nil},
		{"%d %s", []string{"%d", "%s"}},
		{"100%% literal", nil},
		{"%%", nil},
		{"%5.2f", []string{"%5.2f"}},
		{"%[1]d", []string{"%[1]d"}},
		{"%-#7.3v", []string{"%-#7.3v"}},
		{"%d%% complete %s", []string{"%d", "%s"}},
	} {
		t.Run(tc.input, func(t *testing.T) {
			got := extractFormatSpecifiers(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("extractFormatSpecifiers(%q) = %v, want %v", tc.input, got, tc.want)
			}

			for i, spec := range tc.want {
				if got[i] != spec {
					t.Fatalf("extractFormatSpecifiers(%q)[%d] = %q, want %q", tc.input, i, got[i], spec)
				}
			}
		})
	}
}
