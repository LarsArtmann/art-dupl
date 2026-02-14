// Package filter provides smart filtering of auto-generated Go code.
package filter

// FilterOption represents a type of generated code to filter.
type FilterOption string

const (
	// FilterSQLC filters sqlc.dev generated files.
	FilterSQLC FilterOption = "sqlc"

	// FilterTempl filters templ.guide generated files.
	FilterTempl FilterOption = "templ"

	// FilterGoEnum filters go-enum generated files.
	FilterGoEnum FilterOption = "go-enum"

	// FilterAll filters all auto-generated code.
	FilterAll FilterOption = "all"
)

// FilterReason represents the reason a file was filtered.
type FilterReason string

const (
	ReasonSQLC           FilterReason = "sqlc"
	ReasonTempl          FilterReason = "templ"
	ReasonGoEnum         FilterReason = "go-enum"
	ReasonIncludePattern FilterReason = "include-pattern"
	ReasonExcludePattern FilterReason = "exclude-pattern"
	ReasonNotFiltered    FilterReason = "not-filtered"
)

// sqlcFilePatterns contains the standard filename patterns for sqlc.dev generated files.
//
//nolint:gochecknoglobals // Lookup table for sqlc file pattern matching
var sqlcFilePatterns = []string{
	"models.go",
	"querier.go",
	"query.sql.go",
	"batch.go",
}
