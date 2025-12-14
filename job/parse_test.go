package job

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	goFile := setupTestFiles(t)

	fchan := make(chan string, 2)
	fchan <- goFile
	close(fchan)

	schan, _ := Parse(fchan)

	select {
	case seq := <-schan:
		if len(seq) == 0 {
			t.Error("Expected some parsed nodes, got empty sequence")
		}
	case <-time.After(5 * time.Second):
		t.Error("Parse timed out - possible deadlock")
	}
}

func TestParseErrorHandling(t *testing.T) {
	fchan := make(chan string, 1)
	fchan <- "nonexistent_file.go"
	close(fchan)

	schan, _ := Parse(fchan)

	select {
	case seq := <-schan:
		// Should receive something (even if empty) for non-existent file
		_ = seq
	case <-time.After(5 * time.Second):
		t.Error("Parse should handle errors gracefully, not block")
	}
}

func TestParseMultipleFiles(t *testing.T) {
	files := setupMultipleTestFiles(t)

	fchan := make(chan string, 3)
	for _, file := range files {
		fchan <- file
	}
	close(fchan)

	schan, _ := Parse(fchan)

	// Should receive sequences for both files
	count := 0
	for count < 2 {
		select {
		case seq := <-schan:
			if len(seq) == 0 {
				t.Error("Expected parsed nodes for valid Go file")
			}
			count++
		case <-time.After(5 * time.Second):
			t.Error("Parse timed out waiting for multiple files")
			return
		}
	}
}
