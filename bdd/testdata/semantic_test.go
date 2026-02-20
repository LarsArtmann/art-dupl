package testdata

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

// This file contains Ginkgo-style test blocks that are structurally identical
// but semantically different (different method names).
//
// Without semantic detection: These blocks will be flagged as duplicates
// because they have identical AST node type sequences.
//
// With semantic detection: These blocks will NOT be flagged as duplicates
// because the method names are different.

// UserService has tests that use GetByID method
func TestUserService_GetByID(t *testing.T) {
	// Setup
	svc := NewUserService(db)
	ctx := context.Background()

	// Execute
	user, err := svc.GetByID(ctx, 123)

	// Verify
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != 123 {
		t.Errorf("expected ID 123, got %d", user.ID)
	}
}

// OrderService has tests that use GetByID method (different from UserService)
func TestOrderService_GetByID(t *testing.T) {
	// Setup
	svc := NewOrderService(db)
	ctx := context.Background()

	// Execute
	order, err := svc.GetByID(ctx, 456)

	// Verify
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.ID != 456 {
		t.Errorf("expected ID 456, got %d", order.ID)
	}
}

// ProductClient has tests that use Create method
func TestProductClient_Create(t *testing.T) {
	// Setup
	client := NewProductClient(apiKey)
	ctx := context.Background()
	product := &Product{Name: "Test"}

	// Execute
	result, err := client.Create(ctx, product)

	// Verify
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "Test" {
		t.Errorf("expected Name Test, got %s", result.Name)
	}
}

// UserClient has tests that use Create method (different service)
func TestUserClient_Create(t *testing.T) {
	// Setup
	client := NewUserClient(apiKey)
	ctx := context.Background()
	user := &User{Email: "test@example.com"}

	// Execute
	result, err := client.Create(ctx, user)

	// Verify
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Email != "test@example.com" {
		t.Errorf("expected Email test@example.com, got %s", result.Email)
	}
}

// Ginkgo-style tests - these are the main use case for semantic detection

func TestUserHandler(t *testing.T) {
	ginkgo.Describe("UserHandler", func() {
		ginkgo.Context("when getting a user", func() {
			ginkgo.It("should return the user", func() {
				handler := NewUserHandler(service)
				req := httptest.NewRequest("GET", "/users/1", nil)
				rec := httptest.NewRecorder()

				handler.GetUser(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(200))
			})
		})
	})
}

func TestOrderHandler(t *testing.T) {
	ginkgo.Describe("OrderHandler", func() {
		ginkgo.Context("when getting an order", func() {
			ginkgo.It("should return the order", func() {
				handler := NewOrderHandler(service)
				req := httptest.NewRequest("GET", "/orders/1", nil)
				rec := httptest.NewRecorder()

				handler.GetOrder(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(200))
			})
		})
	})
}

func TestProductHandler(t *testing.T) {
	ginkgo.Describe("ProductHandler", func() {
		ginkgo.Context("when getting a product", func() {
			ginkgo.It("should return the product", func() {
				handler := NewProductHandler(service)
				req := httptest.NewRequest("GET", "/products/1", nil)
				rec := httptest.NewRecorder()

				handler.GetProduct(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(200))
			})
		})
	})
}
