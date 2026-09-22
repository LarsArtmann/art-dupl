// Package jsonutil provides the project's canonical JSON marshaling.
//
// art-dupl standardized on encoding/json/v2 output semantics (no HTML
// escaping of <, >, & and no trailing newline) when it migrated to
// GOEXPERIMENT=jsonv2. Go 1.27 removed the v2 struct-tag grammar
// (format:, see go.dev/issue/71631), so the project uses the stable v1 API
// with the v2 engine underneath. v1 defaults differ in exactly two visible
// ways; this package neutralizes both so output bytes stay identical to the
// v2-era wire format that golden files and downstream consumers expect:
//
//   - v1 escapes <, >, & by default; v2 never did.
//   - v1 Encoder.Encode appends a trailing newline; v2 Marshal did not.
package jsonutil

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
)

// MarshalIndent marshals v with the given prefix and indent, without HTML
// escaping and without a trailing newline. An empty indent disables
// indentation. It is the drop-in replacement for the v2-era
// json.Marshal(v, jsontext.WithIndentPrefix(prefix), jsontext.WithIndent(indent)).
func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	var buf bytes.Buffer

	enc := jsontext.NewEncoder(&buf, jsontext.EscapeForHTML(false), jsontext.WithIndentPrefix(prefix), jsontext.WithIndent(indent))

	if err := json.MarshalEncode(enc, v); err != nil {
		return nil, fmt.Errorf("encode json: %w", err)
	}

	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}

// Marshal is MarshalIndent with indentation disabled.
func Marshal(v any) ([]byte, error) {
	return MarshalIndent(v, "", "")
}
