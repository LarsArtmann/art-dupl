package filter

import (
	"github.com/LarsArtmann/gogenfilter"
)

type MetricsMixin = gogenfilter.MetricsMixin
type Metrics = gogenfilter.Metrics
type FilterStats = gogenfilter.FilterStats

func NewMetrics() *Metrics {
	return gogenfilter.NewMetrics()
}
