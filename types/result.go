package types

import (
	"github.com/LarsArtmann/art-dupl/errors"
	"github.com/samber/mo"
)

// Result represents a type-safe operation result.
// This is a thin wrapper around samber/mo.Result for backward compatibility.
type Result[T any] mo.Result[T]

// Ok creates a successful result.
func Ok[T any](value T) Result[T] {
	return Result[T](mo.Ok(value))
}

// Err creates an error result.
func Err[T any](err error) Result[T] {
	return Result[T](mo.Err[T](err))
}

// Errf creates an error result with formatted message.
func Errf[T any](format string, args ...any) Result[T] {
	return Result[T](mo.Errf[T](format, args...))
}

// IsOk checks if result is successful.
func (r Result[T]) IsOk() bool {
	return mo.Result[T](r).IsOk()
}

// IsErr checks if result is an error.
func (r Result[T]) IsErr() bool {
	return mo.Result[T](r).IsError()
}

// Unwrap returns value and error.
func (r Result[T]) Unwrap() (T, error) {
	return mo.Result[T](r).Get()
}

// Value returns the value (for backward compatibility with tests).
// Returns zero value if error.
func (r Result[T]) Value() T {
	val, _ := mo.Result[T](r).Get()
	return val
}

// Err returns the error (for backward compatibility with tests).
// Returns nil if ok.
func (r Result[T]) Err() error {
	return mo.Result[T](r).Error()
}

// Or returns value or default if error.
func (r Result[T]) Or(defaultValue T) T {
	return mo.Result[T](r).OrElse(defaultValue)
}

// OrPanic returns value or panics if error.
func (r Result[T]) OrPanic() T {
	return mo.Result[T](r).MustGet()
}

// Map applies function to result value if ok, otherwise returns error.
// Note: Custom implementation since mo.Result.Map doesn't support type transformation.
func Map[T, U any](r Result[T], fn func(T) U) Result[U] {
	if r.IsErr() {
		return Err[U](mo.Result[T](r).Error())
	}
	val, _ := mo.Result[T](r).Get()
	return Ok(fn(val))
}

// MapErr applies function to error if error, otherwise returns ok.
// Note: Custom implementation to match our API signature.
func (r Result[T]) MapErr(fn func(error) error) Result[T] {
	if r.IsOk() {
		return r
	}
	return Err[T](fn(mo.Result[T](r).Error()))
}

// Validate provides generic validation with proper error integration.
func Validate[T any](value T, validator func(T) bool, errorMsg string) Result[T] {
	if !validator(value) {
		return Err[T](errors.NewValidationError(errorMsg, nil))
	}
	return Ok(value)
}

// Validatef provides generic validation with formatted error.
func Validatef[T any](value T, validator func(T) bool, format string, args ...any) Result[T] {
	if !validator(value) {
		return Err[T](errors.NewValidationError(format, nil))
	}
	return Ok(value)
}

// Option represents optional values with type safety.
// This is a thin wrapper around samber/mo.Option for backward compatibility.
type Option[T any] mo.Option[T]

// Some creates a value present option.
func Some[T any](value T) Option[T] {
	return Option[T](mo.Some(value))
}

// None creates an empty option.
func None[T any]() Option[T] {
	return Option[T](mo.None[T]())
}

// IsSome checks if option has value.
func (o Option[T]) IsSome() bool {
	return mo.Option[T](o).IsPresent()
}

// IsNone checks if option is empty.
func (o Option[T]) IsNone() bool {
	return mo.Option[T](o).IsAbsent()
}

// Unwrap returns value or zero if none.
func (o Option[T]) Unwrap() T {
	val, _ := mo.Option[T](o).Get()
	return val
}

// Or returns value or default if none.
func (o Option[T]) Or(defaultValue T) T {
	return mo.Option[T](o).OrElse(defaultValue)
}

// ToResult converts option to result with error.
func (o Option[T]) ToResult(errorMsg string) Result[T] {
	if o.IsSome() {
		return Ok(o.Unwrap())
	}
	return Err[T](errors.NewValidationError(errorMsg, nil))
}

// Filter applies predicate to option.
func (o Option[T]) Filter(predicate func(T) bool) Option[T] {
	if o.IsSome() && predicate(o.Unwrap()) {
		return o
	}
	return None[T]()
}
