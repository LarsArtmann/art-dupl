package bdd

import (
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// BDD Test Suite for Semantic-Aware Detection
//
// These tests verify the semantic-aware duplicate detection feature,
// which reduces false positives by including identifier names in matching.
//
// The scenarios cover:
// - Structural duplicates detected without semantic flag
// - Semantic-aware detection prevents false positives
// - Config file support for semantic detection
// - Backward compatibility (default: off)

var _ = Describe("Semantic Detection", func() {
	var setup *testutil.BDDTestSetup

	BeforeEach(func() {
		var err error
		setup, err = testutil.NewBDDTestSetupForGinkgo()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(setup.Cleanup()).NotTo(HaveOccurred())
	})

	Context("When semantic detection is disabled (default)", func() {
		It("should detect structural duplicates even with different method names", func() {
			// Create two files with identical AST structure but different method names
			// These should be flagged as duplicates WITHOUT --semantic
			code1 := `package main

import "testing"

func TestUserService_GetByID(t *testing.T) {
	svc := NewUserService(db)
	user, err := svc.GetByID(ctx, 123)
	if err != nil {
		t.Fatal(err)
	}
	if user.ID != 123 {
		t.Error("wrong ID")
	}
}`

			code2 := `package main

import "testing"

func TestOrderService_GetByID(t *testing.T) {
	svc := NewOrderService(db)
	order, err := svc.GetByID(ctx, 456)
	if err != nil {
		t.Fatal(err)
	}
	if order.ID != 456 {
		t.Error("wrong ID")
	}
}`

			err := setup.FileProcessor.WriteFile(filepath.Join(setup.TmpDir, "user_test.go"), []byte(code1), 0o644)
			Expect(err).NotTo(HaveOccurred())
			err = setup.FileProcessor.WriteFile(filepath.Join(setup.TmpDir, "order_test.go"), []byte(code2), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Run WITHOUT --semantic flag (default behavior)
			cmd := exec.Command(setup.BinaryPath, setup.TmpDir, "--threshold", "10")
			output, err := cmd.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			// Should detect structural duplicate
			Expect(outputStr).To(ContainSubstring("user_test.go"))
			Expect(outputStr).To(ContainSubstring("order_test.go"))
		})
	})

	Context("When semantic detection is enabled", func() {
		It("should NOT flag structurally identical code with different method names", func() {
			// Create two files with identical AST structure but ALL different identifiers
			// These should NOT be flagged as duplicates WITH --semantic
			code1 := `package main

import "testing"

func TestUserService_GetByID(tb *testing.T) {
	userSvc := NewUserService(userDB)
	user, userErr := userSvc.GetUser(userCtx, 123)
	if userErr != nil {
		tb.Fatal(userErr)
	}
	if user.ID != 123 {
		tb.Error("wrong user ID")
	}
}`

			code2 := `package main

import "testing"

func TestOrderService_GetByID(tb *testing.T) {
	orderSvc := NewOrderService(orderDB)
	order, orderErr := orderSvc.GetOrder(orderCtx, 456)
	if orderErr != nil {
		tb.Fatal(orderErr)
	}
	if order.ID != 456 {
		tb.Error("wrong order ID")
	}
}`

			err := setup.FileProcessor.WriteFile(filepath.Join(setup.TmpDir, "user_test.go"), []byte(code1), 0o644)
			Expect(err).NotTo(HaveOccurred())
			err = setup.FileProcessor.WriteFile(filepath.Join(setup.TmpDir, "order_test.go"), []byte(code2), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Run WITH --semantic flag
			cmd := exec.Command(setup.BinaryPath, setup.TmpDir, "--threshold", "10", "--semantic")
			output, err := cmd.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			// Should NOT detect as duplicate because ALL identifiers differ
			Expect(outputStr).ToNot(ContainSubstring("user_test.go"))
			Expect(outputStr).ToNot(ContainSubstring("order_test.go"))
		})

		It("should still detect true duplicates when identifiers match", func() {
			// Create two files with identical code (true duplicates)
			code := `package main

import "testing"

func TestSameFunction(t *testing.T) {
	svc := NewService(db)
	result, err := svc.Process(ctx, 123)
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != 123 {
		t.Error("wrong ID")
	}
}`

			err := setup.FileProcessor.WriteDuplicateFiles([]string{"test1_test.go", "test2_test.go"}, code)
			Expect(err).NotTo(HaveOccurred())

			// Run WITH --semantic flag
			cmd := exec.Command(setup.BinaryPath, setup.TmpDir, "--threshold", "10", "--semantic")
			output, err := cmd.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())

			outputStr := string(output)
			// Should still detect true duplicates
			Expect(outputStr).To(ContainSubstring("test1_test.go"))
			Expect(outputStr).To(ContainSubstring("test2_test.go"))
		})
	})

	Context("When using config file with semantic setting", func() {
		It("should respect semantic: true in config file", func() {
			// Create test files with ALL different identifiers
			code1 := `package main

func processUser() {
	userData := getUser()
	userResult := transformUser(userData)
	saveUserResult(userResult)
}`

			code2 := `package main

func processOrder() {
	orderData := getOrder()
	orderResult := transformOrder(orderData)
	saveOrderResult(orderResult)
}`

			err := setup.FileProcessor.WriteFile("user.go", []byte(code1), 0o644)
			Expect(err).NotTo(HaveOccurred())
			err = setup.FileProcessor.WriteFile("order.go", []byte(code2), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Create config file with semantic: true
			configContent := `{
				"semantic": true,
				"threshold": 5
			}`
			configPath := filepath.Join(setup.TmpDir, "dupl.json")
			err = setup.FileProcessor.WriteFile("dupl.json", []byte(configContent), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Run with config file
			cmd := exec.Command(setup.BinaryPath, "-c", configPath, setup.TmpDir)
			output, err := cmd.CombinedOutput()
			outputStr := string(output)
			Expect(err).ToNot(HaveOccurred())

			// Should NOT flag as duplicates due to semantic detection (ALL identifiers differ)
			Expect(outputStr).ToNot(ContainSubstring("user.go"))
			Expect(outputStr).ToNot(ContainSubstring("order.go"))
		})
	})

	Context("When analyzing Ginkgo test patterns", func() {
		It("should distinguish between different handler tests with semantic detection", func() {
			// This is the main use case: Ginkgo test blocks for different handlers
			// have identical structure but test different things
			code1 := `package handler_test

import (
	"net/http/httptest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUserHandler(t *testing.T) {
	Describe("UserHandler", func() {
		Context("when getting a user", func() {
			It("should return the user", func() {
				handler := NewUserHandler(service)
				req := httptest.NewRequest("GET", "/users/1", nil)
				rec := httptest.NewRecorder()

				handler.GetUser(rec, req)

				Expect(rec.Code).To(Equal(200))
			})
		})
	})
}`

			code2 := `package handler_test

import (
	"net/http/httptest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOrderHandler(t *testing.T) {
	Describe("OrderHandler", func() {
		Context("when getting an order", func() {
			It("should return the order", func() {
				handler := NewOrderHandler(service)
				req := httptest.NewRequest("GET", "/orders/1", nil)
				rec := httptest.NewRecorder()

				handler.GetOrder(rec, req)

				Expect(rec.Code).To(Equal(200))
			})
		})
	})
}`

			err := setup.FileProcessor.WriteFile("user_handler_test.go", []byte(code1), 0o644)
			Expect(err).NotTo(HaveOccurred())
			err = setup.FileProcessor.WriteFile("order_handler_test.go", []byte(code2), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Run WITHOUT --semantic: should detect as duplicate
			cmdNoSemantic := exec.Command(setup.BinaryPath, setup.TmpDir, "--threshold", "15")
			outputNoSemantic, err := cmdNoSemantic.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())
			noSemanticStr := string(outputNoSemantic)
			Expect(noSemanticStr).To(ContainSubstring("user_handler_test.go"))

			// Run WITH --semantic: should NOT detect as duplicate
			cmdSemantic := exec.Command(setup.BinaryPath, setup.TmpDir, "--threshold", "15", "--semantic")
			outputSemantic, err := cmdSemantic.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())
			semanticStr := string(outputSemantic)
			// Different handler types should not match (NewUserHandler vs NewOrderHandler, GetUser vs GetOrder)
			Expect(semanticStr).ToNot(ContainSubstring("user_handler_test.go"))
			Expect(semanticStr).ToNot(ContainSubstring("order_handler_test.go"))
		})
	})

	Context("When enum pattern methods have same structure but different receiver types", func() {
		It("should NOT flag methods on different types as duplicates with --semantic", func() {
			// This is the exact pattern from auto-deduplicate that caused false positives
			code1 := `package enums

type CrushMode string

func (cm CrushMode) IsValid() bool {
	return cm != ""
}

func (cm CrushMode) IsEnabled() bool {
	return cm != "disabled"
}

func ParseCrushMode(s string) CrushMode {
	return CrushMode(s)
}`

			code2 := `package enums

type SafetyMode string

func (sm SafetyMode) IsValid() bool {
	return sm != ""
}

func (sm SafetyMode) IsEnabled() bool {
	return sm != "disabled"
}

func ParseSafetyMode(s string) SafetyMode {
	return SafetyMode(s)
}`

			err := setup.FileProcessor.WriteFile("crush_mode.go", []byte(code1), 0o644)
			Expect(err).NotTo(HaveOccurred())
			err = setup.FileProcessor.WriteFile("safety_mode.go", []byte(code2), 0o644)
			Expect(err).NotTo(HaveOccurred())

			// Run WITHOUT --semantic: should detect as duplicate (structural match)
			cmdNoSemantic := exec.Command(setup.BinaryPath, setup.TmpDir, "--threshold", "10")
			outputNoSemantic, err := cmdNoSemantic.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())
			noSemanticStr := string(outputNoSemantic)
			Expect(noSemanticStr).To(ContainSubstring("crush_mode.go"))
			Expect(noSemanticStr).To(ContainSubstring("safety_mode.go"))

			// Run WITH --semantic: should NOT detect as duplicate
			// Because receiver types (CrushMode vs SafetyMode) and function names (ParseCrushMode vs ParseSafetyMode) differ
			cmdSemantic := exec.Command(setup.BinaryPath, setup.TmpDir, "--threshold", "10", "--semantic")
			outputSemantic, err := cmdSemantic.CombinedOutput()
			Expect(err).ToNot(HaveOccurred())
			semanticStr := string(outputSemantic)
			Expect(semanticStr).ToNot(ContainSubstring("crush_mode.go"))
			Expect(semanticStr).ToNot(ContainSubstring("safety_mode.go"))
		})
	})
})
