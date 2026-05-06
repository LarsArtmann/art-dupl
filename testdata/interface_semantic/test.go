package testdata

import "context"

// User represents a user.
type User struct {
	Name string
}

// Service is the interface.
type Service interface {
	GetUser(ctx context.Context, id string) (*User, error)
}

// ServiceImpl is the concrete implementation.
type ServiceImpl struct{}

func (s *ServiceImpl) GetUser(ctx context.Context, id string) (*User, error) {
	if id == "" {
		return nil, context.DeadlineExceeded
	}
	return &User{Name: "test"}, nil
}
