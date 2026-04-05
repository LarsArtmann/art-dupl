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

	// ComponentDeclaration represents a component declaration.
	ComponentDeclaration
	CSSDeclaration
	ScriptDeclaration

	// Element represents an HTML element.
	Element
	TagStart
	TagEnd
	SelfClosingTag
	Doctype

	// StyleElement represents a style element (structural, not content).
	StyleElement
	ScriptElement

	// ComponentIfStatement represents a component if statement.
	ComponentIfStatement
	ComponentForStatement
	ComponentSwitchStatement
	ComponentSwitchExpressionCase
	ComponentSwitchDefaultCase

	// Attribute represents an attribute (structural).
	Attribute
	SpreadAttributes
	ConditionalAttributeIfStatement

	// Expression represents an expression or block.
	Expression
	ComponentBlock
	ComponentRender
	ComponentChildrenExpression
	RawGoBlock

	// ComponentImport represents an import.
	ComponentImport

	// File represents the file root.
	File
)
