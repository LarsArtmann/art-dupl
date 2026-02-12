package domain

import stderrors "errors"

// validationRule represents a single validation rule.
type validationRule struct {
	valid bool
	msg   string
}

// validateRules checks all validation rules and returns the first error encountered.
func validateRules(rules []validationRule) error {
	for _, rule := range rules {
		if !rule.valid {
			return stderrors.New(rule.msg)
		}
	}
	return nil
}

// validateFields is a helper that creates validation rules and validates them in one call.
// This eliminates the repetitive pattern of creating a slice and immediately passing it to validateRules.
func validateFields(rules ...validationRule) error {
	return validateRules(rules)
}
