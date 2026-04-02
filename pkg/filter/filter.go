package filter

import (
	"github.com/LarsArtmann/gogenfilter"
)

type Filter = gogenfilter.Filter

func NewFilter(enabled bool, options []FilterOption) *Filter {
	return gogenfilter.NewFilter(enabled, options)
}
