package config

import "reflect"

// MergeConfigs merges two configurations, with command line config taking precedence.
func MergeConfigs(fileConfig, cliConfig *Config) *Config {
	result := DefaultConfig()

	mergeFileConfig(result, fileConfig)
	mergeCLIConfig(result, cliConfig)

	return result
}

// mergeConfig merges source config into result config.
// If skipZeroValues is true, fields with zero/empty values are skipped.
// Uses reflection to automatically handle all Config fields — adding a new field
// to Config struct is automatically picked up without touching this function.
func mergeConfig(result, cfg *Config, skipZeroValues bool) {
	if cfg == nil {
		return
	}

	resultVal := reflect.ValueOf(result).Elem()
	cfgVal := reflect.ValueOf(cfg).Elem()

	for i := 0; i < cfgVal.NumField(); i++ {
		srcField := cfgVal.Field(i)
		fieldKind := srcField.Kind()

		if skipZeroValues && isFieldZero(srcField, fieldKind) {
			continue
		}

		resultVal.Field(i).Set(srcField)
	}
}

// isFieldZero reports whether a reflect.Value should be treated as "unset"
// for the purpose of config merging.
func isFieldZero(v reflect.Value, kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Slice:
		return v.IsNil() || v.Len() == 0
	default:
		return v.IsZero()
	}
}

func mergeFileConfig(result, cfg *Config) {
	mergeConfig(result, cfg, false)
}

func mergeCLIConfig(result, cfg *Config) {
	mergeConfig(result, cfg, true)
}
