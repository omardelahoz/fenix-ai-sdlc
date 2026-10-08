package grammar

import (
	"github.com/omardelahoz/fenix-ai-sdlc/core/internal/parser/lexer"
)

// Grammar defines the rules for parsing a specific language (FDL, WDL, PMF).
// The Parser Engine is generic and uses a Grammar to understand the syntax.
type Grammar interface {
	// Name returns the name of the grammar (e.g., "FDL", "WDL", "PMF").
	Name() string

	// ParseRoot parses the root of the document.
	ParseRoot(parser *ParserContext) (Node, error)

	// IsKeyword returns true if the given string is a keyword in this grammar.
	IsKeyword(ident string) bool

	// GetKeywordType returns the token type for a keyword.
	GetKeywordType(ident string) lexer.TokenType
}

// ParserContext provides context for parsing.
type ParserContext struct {
	Lexer       *lexer.Lexer
	CurrentToken lexer.Token
	PeekToken   lexer.Token
	Errors      []ParseError
}

// ParseError represents a parsing error.
type ParseError struct {
	Message string
	Span    lexer.Span
}

// Node represents a node in the parse tree.
type Node interface {
	// Type returns the type of the node.
	Type() NodeType

	// Span returns the span of the node in the source.
	Span() lexer.Span

	// Children returns the child nodes.
	Children() []Node
}

// NodeType represents the type of a parse tree node.
type NodeType string

const (
	NodeTypeDocument      NodeType = "Document"
	NodeTypeProduct       NodeType = "Product"
	NodeTypeVision        NodeType = "Vision"
	NodeTypeUsers         NodeType = "Users"
	NodeTypeConstraints   NodeType = "Constraints"
	NodeTypeFeature       NodeType = "Feature"
	NodeTypeStory         NodeType = "Story"
	NodeTypeRequirement   NodeType = "Requirement"
	NodeTypeAcceptance    NodeType = "Acceptance"
	NodeTypeArchitecture  NodeType = "Architecture"
	NodeTypeImplementation NodeType = "Implementation"
	NodeTypeRelease       NodeType = "Release"
	NodeTypeKeyValue      NodeType = "KeyValue"
	NodeTypeString        NodeType = "String"
	NodeTypeNumber        NodeType = "Number"
	NodeTypeBoolean       NodeType = "Boolean"
	NodeTypeList          NodeType = "List"
	NodeTypeBlock         NodeType = "Block"
)

// ParseRule represents a parsing rule.
type ParseRule func(ctx *ParserContext) (Node, error)
