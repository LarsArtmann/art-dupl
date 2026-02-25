package bdd

// Test code samples for semantic detection tests
// Extracted to reduce file size of semantic_detection_test.go

var (
	// structuralTestCode1 and structuralTestCode2 have identical AST structure
	// but different method names - should be flagged with --structural
	structuralTestCode1 = `package main

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

	structuralTestCode2 = `package main

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

	// semanticDifferentCode1 and semanticDifferentCode2 have ALL different identifiers
	// should NOT be flagged with semantic detection
	semanticDifferentCode1 = `package main

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

	semanticDifferentCode2 = `package main

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

	// trueDuplicateCode is identical in both files - should always be detected
	trueDuplicateCode = `package main

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

	// configTestDifferentCode1 and configTestDifferentCode2 test config-based semantic
	configTestDifferentCode1 = `package main

func processUser() {
	userData := getUser()
	userResult := transformUser(userData)
	saveUserResult(userResult)
}`

	configTestDifferentCode2 = `package main

func processOrder() {
	orderData := getOrder()
	orderResult := transformOrder(orderData)
	saveOrderResult(orderResult)
}`

	// handlerTestCode1 and handlerTestCode2 are Ginkgo-style test patterns
	// These have identical structure but test different handlers
	handlerTestCode1 = `package handler_test

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

	handlerTestCode2 = `package handler_test

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

	// enumPatternCode1 and enumPatternCode2 test enum-type method patterns
	// Same structure but different receiver types
	enumPatternCode1 = `package enums

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

	enumPatternCode2 = `package enums

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
)
