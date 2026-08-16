package stats

import (
	"encoding/json/v2"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/printer"
)

func testCacheMetrics() *printer.CacheMetrics {
	return &printer.CacheMetrics{
		Hits:       7,
		Misses:     3,
		MemoryHits: 4,
		Entries:    10,
		HitRatePct: 70,
	}
}

// applyCache simulates the ApplyStatsConfig cache wiring used by the stats
// subcommand.
func applyCache(sp *stats, cache *printer.CacheMetrics) {
	sp.ApplyStatsConfig(printer.StatsConfig{Cache: cache})
}

func TestCacheSectionText(t *testing.T) {
	sp, buf := newTestStatsPrinter()
	applyCache(sp, testCacheMetrics())
	sp.format = config.OutputFormatText

	if err := sp.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	output := buf.String()

	for _, want := range []string{"Cache:", "70.0% (7 hits, 3 misses)", "Memory Hits: 4", "Cached Entries: 10"} {
		if !strings.Contains(output, want) {
			t.Errorf("text output missing %q\noutput:\n%s", want, output)
		}
	}
}

func TestCacheSectionAbsentWithoutIncremental(t *testing.T) {
	sp, buf := newTestStatsPrinter()
	applyCache(sp, nil)
	sp.format = config.OutputFormatText

	if err := sp.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	if strings.Contains(buf.String(), "Cache:") {
		t.Errorf("text output contains cache section for a non-incremental run:\n%s", buf.String())
	}
}

func TestCacheSectionJSON(t *testing.T) {
	sp, buf := newTestStatsPrinter()
	applyCache(sp, testCacheMetrics())
	sp.format = config.OutputFormatJSON

	if err := sp.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	var result struct {
		Cache *printer.CacheMetrics `json:"cache"`
	}

	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}

	if result.Cache == nil {
		t.Fatal("JSON output missing cache object")
	}

	if result.Cache.Hits != 7 || result.Cache.Misses != 3 || result.Cache.MemoryHits != 4 ||
		result.Cache.Entries != 10 || result.Cache.HitRatePct != 70 {
		t.Errorf("JSON cache metrics mismatch: %+v", result.Cache)
	}
}

func TestCacheSectionCSV(t *testing.T) {
	sp, buf := newTestStatsPrinter()
	applyCache(sp, testCacheMetrics())
	sp.format = config.OutputFormatCSV

	if err := sp.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	output := buf.String()

	for _, want := range []string{"Cache Hit Rate,70.0%", "Cache Hits,7", "Cache Misses,3", "Cache Entries,10"} {
		if !strings.Contains(output, want) {
			t.Errorf("CSV output missing %q\noutput:\n%s", want, output)
		}
	}
}
