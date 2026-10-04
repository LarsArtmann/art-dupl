// Package templ provides AST parsing for templ files using the official templ parser.
// It transforms templ template syntax into unified syntax.Node for clone detection.
//
// We focus on structural elements: declarations, HTML elements, flow control, attributes.
// We skip: text content, attribute values, identifiers.
package templ

// Node type constants for templ syntax.
// Meaningful types capture structure without content.
// art-dupl:accept parallel iota enum: independent from syntax/golang/nodetypes.go (different AST domains)
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
	//art-dupl:accept parallel iota enum: name overlaps with golang.File but different domain
	File
)

// Structure types introduced with expression-aware detection (2026-10).
//
// Values sit ABOVE the golang node-type range (0-53) so the serializer's
// IsStatementContainer check can match the else container without colliding
// with any Go node type (golang.Ident is 26, for example).
const (
	// ComponentElseStatement wraps an else branch (if/else bodies and
	// conditional-attribute else lists) so `if X { A } else { B }` differs
	// from `if X { A; B }`. It is a statement CONTAINER (unmarked, like the
	// Go BlockStmt): its statement children are emitted through the
	// IsStatementContainer descent.
	ComponentElseStatement = 100

	// ComponentFallthroughStatement represents a fallthrough in a switch;
	// previously typed as a case body, making it indistinguishable.
	ComponentFallthroughStatement = 101
)
