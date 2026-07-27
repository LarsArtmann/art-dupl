package discordsync

// Finding 4: bool-to-string functions — 3 groups in real codebase.
// Same AST shape, completely different domains and return values.
// Expected actionability: guard-clause (suppressed) when the if-statement
// is isolated from the trailing return.
//
// Fixture design: each function has unique trailing code so the clone detector
// isolates just the guard-clause if-statement.

func fmtBool(b bool) string {
	if b {
		return "Yes"
	}

	return defaultNegative()
}

func emojiExtension(animated bool) string {
	if animated {
		return ".gif"
	}

	return staticExtension()
}

func banActiveSelectValue(activeOnly bool) string {
	if activeOnly {
		return "1"
	}

	return emptyValue()
}
