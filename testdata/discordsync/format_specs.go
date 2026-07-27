package discordsync

// Finding 5: format specifier difference — 1 group in real codebase.
// %06x (lowercase) vs %06X (uppercase) produce different output.
// Expected: NOT suppressed by default (the format specifier IS semantically distinct).
// This is a borderline case — the user accepts it with //art-dupl:accept.

func embedBorderColor(color int) string {
	if color == 0 {
		return ""
	}

	return sprintf("#%06x", color)
}

func roleDotColor(color int) string {
	if color == 0 {
		return ""
	}

	return sprintf("#%06X", color)
}
