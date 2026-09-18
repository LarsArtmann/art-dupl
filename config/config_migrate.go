package config

import (
	"encoding/json"
	"fmt"
)

// UnmarshalJSON implements custom JSON decoding with backward-compatibility
// for the legacy "semantic" bool field that preceded the DetectionMode enum.
//
// Migration rules (only when "detectionMode" is absent from the JSON):
//   - "semantic": false  → DetectionModeExact
//   - "semantic": true   → DetectionModeSemantic (already the default, no-op)
func (c *Config) UnmarshalJSON(data []byte) error {
	type alias Config // break recursion

	a := alias(*c) // preserve pre-set defaults from DefaultConfig()

	if err := json.Unmarshal(data, &a); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	*c = Config(a)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("unmarshal config for migration scan: %w", err)
	}

	if _, hasNew := raw["detectionMode"]; hasNew {
		return nil
	}

	if semanticRaw, hasLegacy := raw["semantic"]; hasLegacy {
		var semantic bool
		if json.Unmarshal(semanticRaw, &semantic) == nil && !semantic {
			c.DetectionMode = DetectionModeExact
		}
	}

	return nil
}
