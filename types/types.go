// Package types provides type-safe functional programming utilities.
//
// This package contains Result[T] and Option[T] types for functional error handling
// and optional value representation. These types are used throughout the codebase
// for operations that may fail or return optional values.
//
// Note: These types are custom implementations. Consider evaluating samber/mo
// for a more feature-complete alternative if additional functional programming
// primitives are needed (flatMap, fold, etc.).
//
// Usage:
//
//	// Result for operations that may fail
//	result := types.Ok(42)
//	if result.IsOk() {
//	    value := result.Unwrap()
//	}
//
//	// Option for optional values
//	opt := types.Some("value")
//	if opt.IsSome() {
//	    value := opt.Unwrap()
//	}
package types
