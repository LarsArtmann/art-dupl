package config

import (
	"fmt"

	"github.com/LarsArtmann/art-dupl/errors"
)

// ValidateConfig validates the configuration.
func ValidateConfig(cfg *Config) error {
	validations := []func() error{
		func() error { return validateThreshold(cfg.Threshold) },
		func() error { return validateMaxChildrenSerial(cfg.MaxChildrenSerial) },
		func() error { return validateOutputFormat(cfg.OutputFormat) },
		func() error { return validateDetectionMethods(cfg.DetectionMethods) },
		func() error { return validateCacheFlags(cfg.CacheDir, cfg.ClearCache, cfg.Incremental) },
		func() error { return validateOnly(cfg.Only) },
	}

	for _, validate := range validations {
		err := validate()
		if err != nil {
			return err
		}
	}

	return nil
}

func validateThreshold(threshold int) error {
	if threshold < 1 {
		return fmt.Errorf("%w: %d", ErrInvalidThreshold, threshold)
	}

	if threshold > 1000 {
		return fmt.Errorf("%w: %d", ErrThresholdTooLarge, threshold)
	}

	return nil
}

func validateMaxChildrenSerial(maxChildren int) error {
	if maxChildren < 1000 {
		return errors.NewValidationError("maxChildrenSerial should be at least 1000", nil)
	}

	if maxChildren > 100000 {
		return errors.NewValidationError("maxChildrenSerial seems too large (max 100000)", nil)
	}

	return nil
}

func validateOutputFormat(format OutputFormat) error {
	if !format.IsValid() {
		return errors.NewValidationError(
			fmt.Sprintf("invalid output format: %s (valid: text, html, json, plumbing)", format),
			nil,
		)
	}

	return nil
}

func validateDetectionMethods(methods []DetectionMethod) error {
	if len(methods) == 0 {
		return errors.NewValidationError("at least one detection method must be specified", nil)
	}

	for _, method := range methods {
		if !method.IsValid() {
			return errors.NewValidationError(
				fmt.Sprintf(
					"invalid detection method: %s (valid: hash, art-dupl, todos, legacy)",
					method,
				),
				nil,
			)
		}
	}

	return nil
}

func validateCacheFlags(cacheDir string, clearCache, incremental bool) error {
	if (cacheDir != "" || clearCache) && !incremental {
		return errors.NewValidationError(
			"--cache-dir and --clear-cache require --incremental mode",
			nil,
		)
	}

	return nil
}

func validateOnly(only FileType) error {
	if !only.IsValid() {
		return errors.NewValidationError(
			fmt.Sprintf("invalid --only value: %q (valid: go, templ)", only),
			nil,
		)
	}

	return nil
}
