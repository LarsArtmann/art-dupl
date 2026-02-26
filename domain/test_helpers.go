package domain

// TestFilename is a reusable test filename constant.
const TestFilename = "test.go"

// MustNewLineNumber creates a LineNumber for tests, panicking on error.
func MustNewLineNumber(n uint16) LineNumber {
	ln, err := NewLineNumber(n)
	if err != nil {
		panic(err)
	}

	return ln
}
