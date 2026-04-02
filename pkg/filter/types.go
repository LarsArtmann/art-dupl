package filter

import (
	"github.com/LarsArtmann/gogenfilter"
)

type FilterOption = gogenfilter.FilterOption
type FilterReason = gogenfilter.FilterReason

const (
	FilterSQLC   = gogenfilter.FilterSQLC
	FilterTempl  = gogenfilter.FilterTempl
	FilterGoEnum = gogenfilter.FilterGoEnum
	FilterAll    = gogenfilter.FilterAll
)

const (
	ReasonSQLC           = gogenfilter.ReasonSQLC
	ReasonTempl          = gogenfilter.ReasonTempl
	ReasonGoEnum         = gogenfilter.ReasonGoEnum
	ReasonIncludePattern = gogenfilter.ReasonIncludePattern
	ReasonExcludePattern = gogenfilter.ReasonExcludePattern
	ReasonNotFiltered    = gogenfilter.ReasonNotFiltered
)
