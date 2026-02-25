// Package templ provides AST parsing for templ files using the official templ parser.
// It transforms templ template syntax into unified syntax.Node for clone detection.
//
// We focus on structural elements: declarations, HTML elements, flow control, attributes.
// We skip: text content, attribute values, identifiers.
package templ

// Node type constants for templ syntax.
// Meaningful types capture structure without content.
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
