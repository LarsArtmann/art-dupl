// Package templ provides AST parsing for templ files using tree-sitter.
//
// This package transforms templ template syntax into the unified syntax.Node
// representation used by art-dupl for clone detection.
//
// Templ is a typed HTML template language for Go that compiles to _templ.go files.
// For clone detection, we focus on structural elements rather than content:
// - Component declarations, CSS declarations, script declarations
// - HTML elements, flow control (if/for/switch)
// - Attributes (structural, not values)
//
// We skip: text content, attribute values, identifiers (names)
package templ

import (
	"os"
	"strings"

	tree_sitter_templ "github.com/LarsArtmann/art-dupl/internal/treesitter/templ"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// Node type constants for templ syntax.
// These map to tree-sitter-templ node types that are meaningful for clone detection.
//
// Meaningful types capture structure without content:
// - Component/CSS/Script declarations define reusable blocks
// - Elements represent HTML structure
// - Flow control affects template structure
// - Attributes capture element properties (not values)
//
// We skip: text content, identifiers, literal values.
const (
	BadNode = iota

	// Core declarations.
	ComponentDeclaration
	CSSDeclaration
	ScriptDeclaration

	// HTML structure.
	Element
	TagStart
	TagEnd
	SelfClosingTag
	Doctype

	// Style/Script elements (structural, not content).
	StyleElement
	ScriptElement

	// Flow control.
	ComponentIfStatement
	ComponentForStatement
	ComponentSwitchStatement
	ComponentSwitchExpressionCase
	ComponentSwitchDefaultCase

	// Attributes (structural).
	Attribute
	SpreadAttributes
	ConditionalAttributeIfStatement

	// Expressions and blocks.
	Expression
	ComponentBlock
	ComponentRender
	ComponentChildrenExpression
	RawGoBlock

	// Imports.
	ComponentImport

	// File root.
	File
)

// nodeTypeMap maps tree-sitter node kind strings to our type constants.
// Only includes node types that are meaningful for clone detection.
var nodeTypeMap = map[string]int32{
	// Core declarations - high value
	"component_declaration": ComponentDeclaration,
	"css_declaration":       CSSDeclaration,
	"script_declaration":    ScriptDeclaration,

	// HTML structure - high value
	"element":          Element,
	"tag_start":        TagStart,
	"tag_end":          TagEnd,
	"self_closing_tag": SelfClosingTag,
	"doctype":          Doctype,

	// Style/Script elements
	"style_element":  StyleElement,
	"script_element": ScriptElement,

	// Flow control - high value
	"component_if_statement":     ComponentIfStatement,
	"component_for_statement":    ComponentForStatement,
	"component_switch_statement": ComponentSwitchStatement,

	// Attributes - structural only
	"attribute":                          Attribute,
	"spread_attributes":                  SpreadAttributes,
	"conditional_attribute_if_statement": ConditionalAttributeIfStatement,

	// Expressions and blocks
	"expression":                    Expression,
	"component_block":               ComponentBlock,
	"component_render":              ComponentRender,
	"component_children_expression": ComponentChildrenExpression,
	"rawgo_block":                   RawGoBlock,

	// Imports
	"component_import": ComponentImport,
}

// skipNodeTypes are node types that should be skipped entirely.
// These are typically content nodes or very granular structural nodes
// that don't contribute meaningfully to clone detection.
var skipNodeTypes = map[string]bool{
	// Content nodes - skip to focus on structure
	"element_text":        true,
	"style_element_text":  true,
	"script_element_text": true,
	"script_block_text":   true,
	"element_comment":     true,

	// Identifier/name nodes - skip to avoid name-based false positives
	"element_identifier":    true,
	"attribute_name":        true,
	"package_identifier":    true,
	"identifier":            true,
	"_component_identifier": true,
	"_css_identifier":       true,
	"_script_identifier":    true,

	// Value nodes - skip to avoid content-based false positives
	"attribute_value":               true,
	"quoted_attribute_value":        true,
	"dynamic_class_attribute_value": true,
	"css_property_value":            true,
	"css_property_name":             true,
	"css_property":                  true,
}

// Parse parses the given templ file and returns the unified syntax tree.
func Parse(filename string) (*syntax.Node, error) {
	node, _, err := ParseWithLineCount(filename)
	return node, err
}

// ParseWithLineCount parses the given templ file and returns the syntax tree along with the line count.
func ParseWithLineCount(filename string) (*syntax.Node, int, error) {
	content, err := os.ReadFile(filename) // #nosec G304 -- filename comes from controlled file system walk
	if err != nil {
		return nil, 0, err //nolint:wrapcheck // Pass through os.ReadFile error
	}

	return ParseBytes(filename, content)
}

// ParseBytes parses templ content and returns the syntax tree along with the line count.
func ParseBytes(filename string, content []byte) (*syntax.Node, int, error) {
	parser := tree_sitter.NewParser()
	defer parser.Close()

	// Set the templ language
	lang := tree_sitter.NewLanguage(tree_sitter_templ.Language())
	if err := parser.SetLanguage(lang); err != nil {
		return nil, 0, err //nolint:wrapcheck // Pass through parser.SetLanguage error
	}

	// Parse the content
	tree := parser.Parse(content, nil)
	if tree == nil {
		return nil, 0, nil
	}
	defer tree.Close()

	root := tree.RootNode()
	if root == nil {
		return nil, 0, nil
	}

	// Transform to unified syntax tree
	t := &transformer{
		content:  content,
		filename: filename,
	}

	node := t.transform(root)

	// Count lines from content
	lineCount := strings.Count(string(content), "\n") + 1

	return node, lineCount, nil
}

type transformer struct {
	content  []byte
	filename string
}

// transform converts a tree-sitter node to a unified syntax.Node.
func (t *transformer) transform(node *tree_sitter.Node) *syntax.Node {
	if node == nil {
		return nil
	}

	kind := node.Kind()

	// Skip node types that shouldn't contribute to clone detection
	if skipNodeTypes[kind] {
		return nil
	}

	// Create the output node
	o := syntax.NewNode()
	o.Filename = t.filename

	// Set positions
	o.Pos = int32(node.StartByte()) // #nosec G115 -- File sizes bounded by int32 in practice
	o.End = int32(node.EndByte())   // #nosec G115 -- File sizes bounded by int32 in practice

	// Map the kind to our type constant
	if typ, ok := nodeTypeMap[kind]; ok {
		o.Type = typ
	} else {
		// For unmapped types, use BadNode but still process children
		o.Type = BadNode
	}

	// Process children (only named children for cleaner AST)
	childCount := node.NamedChildCount()
	for i := range childCount {
		child := node.NamedChild(i)
		if child == nil {
			continue
		}
		childNode := t.transform(child)
		if childNode != nil {
			o.AddChildren(childNode)
		}
	}

	return o
}
