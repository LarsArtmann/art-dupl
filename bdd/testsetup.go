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
// Usage:
//
//	var setup *testutil.BDDTestSetup
//	BeforeEach(func() {
//	    setup = bdd.CreateBDDTestSetup()
//	})
//
//	It("should do something", func() {
//	    // use setup...
//	})
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

	return setup
}
