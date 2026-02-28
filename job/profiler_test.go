package job

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestProfile(t *testing.T) {
	// Test that Profile() captures metrics without panicking
	result := Profile()

	if result.AllocMB < 0 {
		t.Errorf("AllocMB should be >= 0, got %f", result.AllocMB)
	}

	if result.NumGoroutine < 0 {
		t.Errorf("NumGoroutine should be >= 0, got %d", result.NumGoroutine)
	}
}

func TestStartProfile(t *testing.T) {
	start := StartProfile()

	if start.Duration != 0 {
		t.Errorf("StartProfile duration should be 0 initially, got %v", start.Duration)
	}

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	end := EndProfile(start)

	if end.Duration == 0 {
		t.Errorf("EndProfile should calculate duration, got 0")
	}
}

func TestProfileWithDuration(t *testing.T) {
	duration := 100 * time.Millisecond
	result := ProfileWithDuration(duration)

	if result.Duration != duration {
		t.Errorf(
			"ProfileWithDuration duration should match input, got %v want %v",
			result.Duration,
			duration,
		)
	}
}

func TestProfileDiff(t *testing.T) {
	start := Profile()

	time.Sleep(10 * time.Millisecond)

	end := Profile()

	diff := ProfileDiff(start, end)

	// Diff should have positive duration (approximately)
	if diff.Duration < 0 {
		t.Errorf("ProfileDiff duration should be >= 0, got %v", diff.Duration)
	}

	// NumGC in diff represents the GC count increase
	// Since it's uint32, it's always >= 0, but we verify it's reasonable (< 1000 GCs)
	if diff.NumGC > 1000 {
		t.Errorf("ProfileDiff NumGC seems too high, got %d", diff.NumGC)
	}
}

func TestContextTimeout(t *testing.T) {
	// Test that timeout context works correctly
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Millisecond)
	defer cancel()

	select {
	case <-time.After(5 * time.Millisecond):
		// Should complete before timeout
	case <-ctx.Done():
		t.Error("Should not timeout yet")
	}
}

func TestContextTimeoutExpired(t *testing.T) {
	// Test that timeout context expires correctly
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Millisecond)
	defer cancel()

	select {
	case <-time.After(10 * time.Millisecond):
		// Should timeout after 5ms
	case <-ctx.Done():
		// Expected timeout
	}

	// Context should be cancelled
	if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Errorf("Context should have deadline exceeded error, got %v", ctx.Err())
	}
}
