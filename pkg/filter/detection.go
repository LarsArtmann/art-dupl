package filter

import (
	"github.com/LarsArtmann/gogenfilter"
)

func matchPattern(path, pattern string) bool {
	return gogenfilter.MatchPattern(path, pattern)
}

func isSQLCGenerated(filePath, content string) bool {
	return gogenfilter.IsSQLCGenerated(filePath, content)
}

func isTemplGenerated(filePath, content string) bool {
	return gogenfilter.IsTemplGenerated(filePath, content)
}

func isGoEnumGenerated(filePath, content string) bool {
	return gogenfilter.IsGoEnumGenerated(filePath, content)
}

func matchesSQLCFilename(filePath string) bool {
	return gogenfilter.MatchesSQLCFilename(filePath)
}

func hasSQLCContent(content string) bool {
	return gogenfilter.HasSQLCContent(content)
}

func hasSQLCCodePatterns(content string) bool {
	return gogenfilter.HasSQLCCodePatterns(content)
}
