package types

import (
	"github.com/LarsArtmann/art-dupl/errors"
)

// Result represents a type-safe operation result
type Result[T any] struct {
	Value T
	Error error
}

// Ok creates a successful result
func Ok[T any](value T) Result[T] {
	return Result[T]{Value: value, Error: nil}
}

// Err creates an error result
func Err[T any](err error) Result[T] {
	return Result[T]{Error: err}
}

// Errf creates an error result with formatted message
func Errf[T any](format string, args ...any) Result[T] {
	return Err[T](errors.NewValidationError(format, nil))
}

// IsOk checks if result is successful
func (r Result[T]) IsOk() bool {
	return r.Error == nil
}

// IsErr checks if result is an error
func (r Result[T]) IsErr() bool {
	return r.Error != nil
}

// Unwrap returns value and error
func (r Result[T]) Unwrap() (T, error) {
	return r.Value, r.Error
}

// Or returns value or default if error
func (r Result[T]) Or(defaultValue T) T {
	if r.Error != nil {
		return defaultValue
	}
	return r.Value
}

// OrPanic returns value or panics if error
func (r Result[T]) OrPanic() T {
	if r.Error != nil {
		panic(r.Error)
	}
	return r.Value
}

// Map applies function to result value if ok, otherwise returns error
func Map[T, U any](r Result[T], fn func(T) U) Result[U] {
	if r.Error != nil {
		return Err[U](r.Error)
	}
	return Ok(fn(r.Value))
}

// MapErr applies function to error if error, otherwise returns ok
func (r Result[T]) MapErr(fn func(error) error) Result[T] {
	if r.Error == nil {
		return r
	}
	return Err[T](fn(r.Error))
}

// Validate provides generic validation with proper error integration
func Validate[T any](value T, validator func(T) bool, errorMsg string) Result[T] {
	if !validator(value) {
		return Err[T](errors.NewValidationError(errorMsg, nil))
	}
	return Ok(value)
}

// Validatef provides generic validation with formatted error
func Validatef[T any](value T, validator func(T) bool, format string, args ...any) Result[T] {
	if !validator(value) {
		return Err[T](errors.NewValidationError(format, nil))
	}
	return Ok(value)
}

// Option represents optional values with type safety
type Option[T any] struct {
	value T
	some  bool
}

// Some creates a value present option
func Some[T any](value T) Option[T] {
	return Option[T]{value: value, some: true}
}

// None creates an empty option
func None[T any]() Option[T] {
	var zero T
	return Option[T]{value: zero, some: false}
}

// IsSome checks if option has value
func (o Option[T]) IsSome() bool {
	return o.some
}

// IsNone checks if option is empty
func (o Option[T]) IsNone() bool {
	return !o.some
}

// Unwrap returns value or zero if none
func (o Option[T]) Unwrap() T {
	return o.value
}

// Or returns value or default if none
func (o Option[T]) Or(defaultValue T) T {
	if o.some {
		return o.value
	}
	return defaultValue
}

// ToResult converts option to result with error
func (o Option[T]) ToResult(errorMsg string) Result[T] {
	if o.some {
		return Ok(o.value)
	}
	return Err[T](errors.NewValidationError(errorMsg, nil))
}

// Filter applies predicate to option
func (o Option[T]) Filter(predicate func(T) bool) Option[T] {
	if o.some && predicate(o.value) {
		return o
	}
	return None[T]()
}
