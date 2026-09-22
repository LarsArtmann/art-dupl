// Package jsonutil provides the project's canonical JSON marshaling.
//
// art-dupl standardized on v2-era output semantics (no HTML escaping of
// <, >, & and no trailing newline) when it migrated to GOEXPERIMENT=jsonv2.
// Go 1.27 removed the v2 struct-tag grammar (format:, go.dev/issue/71631),
// and the pure encoding/json/v2 API also refuses plain time.Duration values
// ("no default representation"), so this package marshals through the stable
// v1 API exclusively. v1 defaults differ from the v2-era wire format in
// exactly two visible ways; both are neutralized here so output bytes stay
// identical to what golden files and downstream consumers expect:
//
//   - v1 escapes <, >, & by default; v2 never did → Encoder.SetEscapeHTML(false).
//   - v1 Encoder.Encode appends a trailing newline; v2 Marshal did not →
//     trimmed here.
//
// v1 defaults format time.Duration as integer nanoseconds, which the config
// save/load round-trip relies on; do not reintroduce direct v2 imports.
package jsonutil

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// MarshalIndent marshals v with the given prefix and indent, without HTML
// escaping and without a trailing newline. An empty indent disables
// indentation. It is the drop-in replacement for the v2-era
// json.Marshal(v, jsontext.WithIndentPrefix(prefix), jsontext.WithIndent(indent)).
func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if indent != "" || prefix != "" {
		enc.SetIndent(prefix, indent)
	}

	if err := enc.Encode(v); err != nil {
		return nil, fmt.Errorf("encode json: %w", err)
	}

	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}

// Marshal is MarshalIndent with indentation disabled.
func Marshal(v any) ([]byte, error) {
	return MarshalIndent(v, "", "")
}
