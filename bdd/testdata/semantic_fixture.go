package testdata

import "context"

// This file contains code blocks that are structurally identical
// but semantically different (different method names).
//
// Without semantic detection: These blocks will be flagged as duplicates
// because they have identical AST node type sequences.
//
// With semantic detection: These blocks will NOT be flagged as duplicates
// because the method names are different.

// UserService has methods that use GetByID
type UserService struct{}

func (s *UserService) GetByID(ctx context.Context, id int) error {
	if id == 0 {
		return nil
	}
	return nil
}

func (s *UserService) ProcessUser(ctx context.Context, id int) error {
	if id == 0 {
		return nil
	}
	return nil
}

// OrderService has methods that use GetByID (different from UserService)
type OrderService struct{}

func (s *OrderService) GetByID(ctx context.Context, id int) error {
	if id == 0 {
		return nil
	}
	return nil
}

func (s *OrderService) ProcessOrder(ctx context.Context, id int) error {
	if id == 0 {
		return nil
	}
	return nil
}

// ProductClient has methods that use Create
type ProductClient struct{}

func (c *ProductClient) Create(ctx context.Context, name string) error {
	if name == "" {
		return nil
	}
	return nil
}

func (c *ProductClient) Update(ctx context.Context, name string) error {
	if name == "" {
		return nil
	}
	return nil
}

// UserClient has methods that use Create (different from ProductClient)
type UserClient struct{}

func (c *UserClient) Create(ctx context.Context, email string) error {
	if email == "" {
		return nil
	}
	return nil
}

func (c *UserClient) Update(ctx context.Context, email string) error {
	if email == "" {
		return nil
	}
	return nil
}

// Handler tests - these are the main use case for semantic detection

type UserHandler struct{}

func (h *UserHandler) GetUser(id int) int {
	if id == 0 {
		return 0
	}
	return id
}

func (h *UserHandler) CreateUser(name string) string {
	if name == "" {
		return ""
	}
	return name
}

type OrderHandler struct{}

func (h *OrderHandler) GetOrder(id int) int {
	if id == 0 {
		return 0
	}
	return id
}

func (h *OrderHandler) CreateOrder(product string) string {
	if product == "" {
		return ""
	}
	return product
}

type ProductHandler struct{}

func (h *ProductHandler) GetProduct(id int) int {
	if id == 0 {
		return 0
	}
	return id
}

func (h *ProductHandler) CreateProduct(price float64) float64 {
	if price == 0 {
		return 0
	}
	return price
}
