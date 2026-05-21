package bdd

import (
	"fmt"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/onsi/ginkgo/v2"
)

// CreateBDDTestSetup creates a BDDTestSetup for use in Ginkgo BeforeEach blocks.
// It automatically registers cleanup with Ginkgo's DeferCleanup, so you only
// need to call this in BeforeEach - no AfterEach needed.
//
// Execution is done in-process (no binary building) for speed and reliability.
func CreateBDDTestSetup() *testutil.BDDTestSetup {
	setup, err := testutil.NewBDDTestSetupForGinkgo()

	ginkgoFail := func(msg string) {
		ginkgo.Fail(msg)
	}
	if err != nil {
		ginkgoFail(fmt.Sprintf("Failed to create BDD test setup: %v", err))
	}

	ginkgo.DeferCleanup(func() {
		cleanupErr := setup.Cleanup()
		if cleanupErr != nil {
			ginkgoFail(fmt.Sprintf("Failed to cleanup BDD test setup: %v", cleanupErr))
		}
	})

	// Wire in-process execution (combined output for backward compatibility)
	setup.Executor = func(args ...string) ([]byte, error) {
		result, err := executeInProcess(args...)
		if err != nil {
			return result.Combined(), err
		}

		return result.Combined(), nil
	}

	// Wire in-process execution (separated stdout/stderr for JSON parsing)
	setup.ExecutorResult = executeInProcess

	return setup
}
