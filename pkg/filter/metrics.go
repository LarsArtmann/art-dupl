package filter

import (
	"github.com/LarsArtmann/gogenfilter"
)

type (
	MetricsMixin = gogenfilter.MetricsMixin
	Metrics      = gogenfilter.Metrics
	FilterStats  = gogenfilter.FilterStats
)

func NewMetrics() *Metrics {
	return gogenfilter.NewMetrics()
}
