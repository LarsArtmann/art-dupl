package discordsync

// Finding 4: bool-to-string functions — 3 groups in real codebase.
// Same AST shape, completely different domains and return values.
// Expected actionability: single-simple-statement or signature-only (suppressed).

func fmtBool(b bool) string {
	if b {
		return "Yes"
	}

	return "No"
}

func emojiExtension(animated bool) string {
	if animated {
		return ".gif"
	}

	return ".png"
}

func banActiveSelectValue(activeOnly bool) string {
	if activeOnly {
		return "1"
	}

	return ""
}
