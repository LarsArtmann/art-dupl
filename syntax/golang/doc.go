// Package golang provides Go AST parsing and transformation for clone detection.
//
// Implementation split across:
// - nodetypes.go: AST node type constants
// - parse.go: Parse functions and transformer struct
// - transform.go: AST transformation logic
// - identifier_hash.go: semantic identifier/operator hashing
// - detection_mode.go: detection mode types
// - parse_config.go: parse configuration
package golang
