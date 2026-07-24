package cmd

import (
	"context"
	"os"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
)

func TestShouldShowProgress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cfg      *config.Config
		format   config.OutputFormat
		expected bool
	}{
		{
			name:     "text output shows progress",
			cfg:      &config.Config{},
			format:   config.OutputFormatText,
			expected: true,
		},
		{
			name:     "verbose shows progress even for JSON",
			cfg:      &config.Config{Verbose: true},
			format:   config.OutputFormatJSON,
			expected: true,
		},
		{
			name:     "quiet suppresses progress",
			cfg:      &config.Config{Quiet: true},
			format:   config.OutputFormatText,
			expected: false,
		},
		{
			name:     "JSON output suppresses progress",
			cfg:      &config.Config{},
			format:   config.OutputFormatJSON,
			expected: false,
		},
		{
			name:     "plumbing output suppresses progress",
			cfg:      &config.Config{},
			format:   config.OutputFormatPlumbing,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := shouldShowProgress(tt.cfg, tt.format)
			if got != tt.expected {
				t.Errorf("shouldShowProgress() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestShouldShowProgressEnvVar(t *testing.T) {
	t.Run("ARTDUPL_NO_PROGRESS=1 suppresses", func(t *testing.T) {
		t.Setenv("ARTDUPL_NO_PROGRESS", "1")

		cfg := &config.Config{}

		if shouldShowProgress(cfg, config.OutputFormatText) {
			t.Error("expected progress to be suppressed by ARTDUPL_NO_PROGRESS=1")
		}
	})

	t.Run("ARTDUPL_NO_PROGRESS unset shows progress", func(t *testing.T) {
		t.Setenv("ARTDUPL_NO_PROGRESS", "")

		cfg := &config.Config{}

		if !shouldShowProgress(cfg, config.OutputFormatText) {
			t.Error("expected progress to show when ARTDUPL_NO_PROGRESS is unset")
		}
	})
}

func TestProgressFilesChanForwarding(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{}
	format := config.OutputFormatText

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	t.Cleanup(func() { os.Stderr = oldStderr })

	input := make(chan string, 3)

	input <- "file1.go"

	input <- "file2.go"

	input <- "file3.go"

	close(input)

	output := progressFilesChan(ctx, input, cfg, format)

	var results []string
	for f := range output {
		results = append(results, f)
	}

	_ = w.Close()

	// drain pipe to avoid goroutine leak
	go func() {
		buf := make([]byte, 1024)

		for {
			if _, err := r.Read(buf); err != nil {
				return
			}
		}
	}()

	if len(results) != 3 {
		t.Errorf("expected 3 files forwarded, got %d", len(results))
	}

	if results[0] != "file1.go" {
		t.Errorf("expected file1.go, got %s", results[0])
	}
}

func TestProgressFilesChanSuppressed(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{Quiet: true}

	input := make(chan string, 2)
	input <- "a.go"

	input <- "b.go"

	close(input)

	output := progressFilesChan(ctx, input, cfg, config.OutputFormatText)

	// When suppressed, progressFilesChan returns the input channel directly
	if output != input {
		// It may have been wrapped; verify files still come through
		count := 0
		for range output {
			count++
		}

		if count != 2 {
			t.Errorf("expected 2 files, got %d", count)
		}
	}
}

func TestProgressFilesChanEmpty(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{Quiet: true}

	input := make(chan string)

	close(input)

	output := progressFilesChan(ctx, input, cfg, config.OutputFormatText)

	count := 0
	for range output {
		count++
	}

	if count != 0 {
		t.Errorf("expected 0 files, got %d", count)
	}
}
