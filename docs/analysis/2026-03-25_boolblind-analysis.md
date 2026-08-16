# BoolBlind Analysis Report: art-dupl

**Analysis Date:** March 25, 2026\
**Tool:** branching-flow v1.0\
**Command:** `branching-flow boolblind .`

---

## Findings Summary

| Severity    | Count | Struct        | Location            | Bool Fields |
| ----------- | ----- | ------------- | ------------------- | ----------- |
| 🚨 Critical | 1     | Config        | config/config.go:60 | 11 bools    |
| ⚠️ High      | 1     | RuntimeConfig | cli/runtime.go:17   | 7 bools     |

---

## Detailed Analysis

### 1. Config struct (config/config.go)

**Fields identified:**

- `IncludeVendor`
- `IncludeNodeModules`
- `FilesFromStdin`
- `Verbose`
- `Profile`
- `FilterGenerated`
- `IncludeSQLC`
- `IncludeTempl`
- `Incremental`
- `ClearCache`
- `Semantic`

**Why bitflags are NOT recommended:**

1. **JSON Serialization Required**: The Config struct is used for JSON/YAML config file loading. Bitflags would require custom marshaling/unmarshaling logic, significantly complicating the config system.

2. **Code Clarity Impact**: The bool fields are used in conditional checks throughout the codebase (100+ references found). Converting to bitflags would require helper methods like `HasFlag(ConfigFlagVerbose)` which reduces readability.

3. **API Breaking Changes**: Changing to bitflags would break existing JSON configs that users have created.

4. **Minimal Memory Savings**: 11 bools = 11 bytes. Bitflags = ~2 bytes. Savings of 9 bytes per instance is negligible for a CLI tool that typically has only 1-2 Config instances.

5. **Test Complexity**: All test code that sets individual bools would need to be refactored.

### 2. RuntimeConfig struct (cli/runtime.go)

**Fields identified:**

- `Vendor`
- `Verbose`
- `FilesFromStdin`
- `HTML`
- `JSON`
- `Plumbing`
- `DiffMode`

**Why bitflags are NOT recommended:**

1. **CLI Flag Mapping**: These bools directly map to CLI flags. Using bitflags would require translation layers between flag parsing and struct fields.

2. **Short-Lived Instances**: RuntimeConfig is created once at startup and converted to Config. Memory savings are irrelevant.

3. **Type Conversion Complexity**: RuntimeConfig has a `ToConfig()` method that maps to config.Config. Adding bitflags would require additional conversion logic.

---

## Recommendation: **DO NOT IMPLEMENT**

Converting these bool fields to bitflags would:

| Aspect               | Impact                   | Assessment   |
| -------------------- | ------------------------ | ------------ |
| Memory savings       | ~9-15 bytes per instance | Negligible   |
| Code complexity      | High increase            | Bad tradeoff |
| Readability          | Decreased                | Negative     |
| Config compatibility | Breaking change          | Risky        |
| Test maintenance     | High burden              | Unnecessary  |
| Performance          | No measurable gain       | Irrelevant   |

### When Bitflags ARE Appropriate

Bitflags make sense when:

1. Storing many boolean states in a database or network protocol
2. Working with hardware registers or binary protocols
3. Need atomic operations on multiple flags
4. Have thousands of instances in memory-constrained environments

None of these apply to art-dupl's Config or RuntimeConfig structs.

---

## Conclusion

The `branching-flow boolblind` analysis correctly identified structs with multiple bool fields, but **applying bitflags here would be premature optimization**. The costs (complexity, reduced readability, breaking changes) far outweigh the minimal benefits (negligible memory savings).

**Recommended Action:** Document this decision and close as "won't fix - not appropriate for this use case."

---

## Verification

To reproduce this analysis:

```bash
# Install branching-flow (if not already installed)
go install github.com/LarsArtmann/branching-flow@latest

# Run the analysis
branching-flow boolblind .

# View all analyzers
branching-flow stats .
```

The analysis correctly identified the structs but the fix is not appropriate for this codebase's requirements.
