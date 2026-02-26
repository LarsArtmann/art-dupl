package config

import (
	"testing"
)

// AssertMergeConfigsWithNil is a helper function for testing merge configs with nil parameters.
func AssertMergeConfigsWithNil(
	t *testing.T,
	testConfig *Config,
	isNilFileConfig bool,
	expectedValues map[string]any,
) {
	t.Helper()

	var (
		merged   *Config
		testName string
	)

	if isNilFileConfig {
		testName = "NilFileConfig"
		merged = MergeConfigs(nil, testConfig)
	} else {
		testName = "NilCLIConfig"
		merged = MergeConfigs(testConfig, nil)
	}

	// Test expected values
	if threshold, ok := expectedValues["threshold"].(int); ok {
		if merged.Threshold != threshold {
			t.Errorf(
				"%s: Expected merged threshold %d, got %d",
				testName,
				threshold,
				merged.Threshold,
			)
		}
	}

	if outputFormat, ok := expectedValues["outputFormat"].(string); ok {
		if merged.OutputFormat.String() != outputFormat {
			t.Errorf(
				"%s: Expected merged OutputFormat %s, got %s",
				testName,
				outputFormat,
				merged.OutputFormat.String(),
			)
		}
	}

	if verbose, ok := expectedValues["verbose"].(bool); ok {
		if merged.Verbose != verbose {
			t.Errorf("%s: Expected merged Verbose %v, got %v", testName, verbose, merged.Verbose)
		}
	}

	if includeVendor, ok := expectedValues["includeVendor"].(bool); ok {
		if merged.IncludeVendor != includeVendor {
			t.Errorf(
				"%s: Expected merged IncludeVendor %v, got %v",
				testName,
				includeVendor,
				merged.IncludeVendor,
			)
		}
	}
}
