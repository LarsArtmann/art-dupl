# ADR 0003: Reflection-Based Config Merge

## Status

Accepted

## Context

art-dupl supports layered configuration: default values → JSON config file → CLI flags. The `mergeConfig()` function was responsible for merging JSON config into the CLI-overridden config.

The implementation was a 170-line manual field-by-field merge:

```go
func mergeConfig(cfg *Config, fileCfg *Config) {
    if !cfg.Changed("threshold") && fileCfg.Threshold != 0 {
        cfg.Threshold = fileCfg.Threshold
    }
    if !cfg.Changed("format") && fileCfg.Format != "" {
        cfg.Format = fileCfg.Format
    }
    // ... repeated for every field
}
```

This approach had two problems:

1. **Maintenance burden**: Every new Config field required adding a merge case
2. **Bug-prone**: Fields could be forgotten, leading to silent merge failures

## Decision

Replaced the 170-line manual merge with a 30-line reflection-based merge that iterates over all struct fields and applies the same logic automatically:

```go
func mergeConfig(cfg *Config, fileCfg *Config) {
    cfgVal := reflect.ValueOf(cfg).Elem()
    fileVal := reflect.ValueOf(fileCfg).Elem()

    for i := range cfgVal.NumField() {
        field := cfgVal.Type().Field(i)
        if !isExported(field) { continue }
        // Apply Changed() check, then copy from fileCfg
    }
}
```

## Consequences

### Positive

- Adding new Config fields no longer requires touching merge code
- Zero risk of forgetting a field
- Reduced from 170 to 30 lines
- Merge logic is consistent across all fields

### Negative

- Reflection is slower than direct field access, but config merge runs once at startup (negligible)
- Less obvious what happens at a glance, developers need to understand the reflection pattern
- Edge cases with non-comparable types require special handling

### Mitigation

The reflection logic is well-tested via existing BDD configuration tests that exercise all merge paths.
