package main

import (
	"testing"
)

func TestCLIInterface(t *testing.T) {
	// Test that RealCLI works correctly
	// Note: We can't test os.Exit directly, but we can verify the interface
	realCLI := &RealCLI{}
	
	// Test that interface is satisfied
	var _ CLIInterface = realCLI
	
	// Test writers
	if realCLI.Stderr() == nil {
		t.Error("Stderr should not be nil")
	}
	if realCLI.Stdout() == nil {
		t.Error("Stdout should not be nil")
	}
}

func TestTestCLI(t *testing.T) {
	testCLI := &TestCLI{}
	
	// Test that interface is satisfied
	var _ CLIInterface = testCLI
	
	// Test Exit code capture
	testCLI.Exit(42)
	if testCLI.ExitCode != 42 {
		t.Errorf("Expected exit code 42, got %d", testCLI.ExitCode)
	}
	
	// Test writer functionality
	testCLI.Stderr().Write([]byte("stderr"))
	testCLI.Stdout().Write([]byte("stdout"))
	
	if testCLI.StderrBuf.String() != "stderr" {
		t.Errorf("Expected stderr buffer content 'stderr', got '%s'", testCLI.StderrBuf.String())
	}
	
	if testCLI.StdoutBuf.String() != "stdout" {
		t.Errorf("Expected stdout buffer content 'stdout', got '%s'", testCLI.StdoutBuf.String())
	}
}

func TestCLISetup(t *testing.T) {
	// Test that we can set up CLI for testing
	originalCLI := cli
	defer func() { cli = originalCLI }()
	
	testCLI := &TestCLI{}
	cli = testCLI
	
	// Verify that cli is now the test implementation
	if cli != testCLI {
		t.Error("Failed to set CLI to test implementation")
	}
}