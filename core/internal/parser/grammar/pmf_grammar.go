package grammar

import (
	"github.com/omardelahoz/fenix-ai-sdlc/core/internal/parser/lexer"
)

// PMFGrammar implements the Grammar interface for PMF (Processor Manifest Format).
type PMFGrammar struct{}

// NewPMFGrammar creates a new PMF grammar.
func NewPMFGrammar() *PMFGrammar {
	return &PMFGrammar{}
}

// Name returns the name of the grammar.
func (g *PMFGrammar) Name() string {
	return "PMF"
}

// ParseRoot parses the root of a PMF document.
func (g *PMFGrammar) ParseRoot(ctx *ParserContext) (Node, error) {
	// PMF document starts with "processor_manifest <name>"
	if !g.expectKeyword(ctx, "processor_manifest") {
		return nil, ctx.Errors[0]
	}

	// Processor manifest name
	if !g.expect(ctx, lexer.TokenIdentifier) {
		return nil, ctx.Errors[0]
	}

	// Create document node
	doc := &ProcessorManifestNode{
		Name: ctx.CurrentToken.Value,
		Span: ctx.CurrentToken.Span,
	}

	// Parse manifest sections
	for !g.check(ctx, lexer.TokenEOF) {
		switch {
		case g.checkKeyword(ctx, "version"):
			version, err := g.parseVersion(ctx)
			if err != nil {
				return nil, err
			}
			doc.Version = version
		case g.checkKeyword(ctx, "author"):
			author, err := g.parseAuthor(ctx)
			if err != nil {
				return nil, err
			}
			doc.Author = author
		case g.checkKeyword(ctx, "description"):
			description, err := g.parseDescription(ctx)
			if err != nil {
				return nil, err
			}
			doc.Description = description
		case g.checkKeyword(ctx, "input"):
			input, err := g.parseInput(ctx)
			if err != nil {
				return nil, err
			}
			doc.Inputs = append(doc.Inputs, input)
		case g.checkKeyword(ctx, "output"):
			output, err := g.parseOutput(ctx)
			if err != nil {
				return nil, err
			}
			doc.Outputs = append(doc.Outputs, output)
		case g.checkKeyword(ctx, "capability"):
			capability, err := g.parseCapability(ctx)
			if err != nil {
				return nil, err
			}
			doc.Capabilities = append(doc.Capabilities, capability)
		default:
			// Skip unknown tokens
			ctx.CurrentToken = ctx.PeekToken
			ctx.PeekToken = ctx.Lexer.NextToken()
		}
	}

	return doc, nil
}

// IsKeyword returns true if the identifier is a PMF keyword.
func (g *PMFGrammar) IsKeyword(ident string) bool {
	switch ident {
	case "processor_manifest", "version", "author", "description",
		"input", "output", "capability", "type", "required", "optional":
		return true
	default:
		return false
	}
}

// GetKeywordType returns the token type for a keyword.
func (g *PMFGrammar) GetKeywordType(ident string) lexer.TokenType {
	switch ident {
	case "processor_manifest":
		return lexer.TokenProcessorManifest
	case "version":
		return lexer.TokenVersion
	case "author":
		return lexer.TokenAuthor
	case "description":
		return lexer.TokenDescription
	case "input":
		return lexer.TokenInput
	case "output":
		return lexer.TokenOutput
	case "capability":
		return lexer.TokenCapability
	default:
		return lexer.TokenIdentifier
	}
}

// Helper methods for parsing PMF structures

func (g *PMFGrammar) expectKeyword(ctx *ParserContext, keyword string) bool {
	if ctx.CurrentToken.Type == lexer.TokenKeyword && ctx.CurrentToken.Value == keyword {
		ctx.CurrentToken = ctx.PeekToken
		ctx.PeekToken = ctx.Lexer.NextToken()
		return true
	}
	ctx.Errors = append(ctx.Errors, ParseError{
		Message: "expected keyword '" + keyword + "'",
		Span:    ctx.CurrentToken.Span,
	})
	return false
}

func (g *PMFGrammar) expect(ctx *ParserContext, tokenType lexer.TokenType) bool {
	if ctx.CurrentToken.Type == tokenType {
		ctx.CurrentToken = ctx.PeekToken
		ctx.PeekToken = ctx.Lexer.NextToken()
		return true
	}
	ctx.Errors = append(ctx.Errors, ParseError{
		Message: "expected " + tokenType.String(),
		Span:    ctx.CurrentToken.Span,
	})
	return false
}

func (g *PMFGrammar) check(ctx *ParserContext, tokenType lexer.TokenType) bool {
	return ctx.CurrentToken.Type == tokenType
}

func (g *PMFGrammar) checkKeyword(ctx *ParserContext, keyword string) bool {
	return ctx.CurrentToken.Type == lexer.TokenKeyword && ctx.CurrentToken.Value == keyword
}

func (g *PMFGrammar) parseVersion(ctx *ParserContext) (*VersionNode, error) {
	g.expectKeyword(ctx, "version")

	version := &VersionNode{
		Span: ctx.CurrentToken.Span,
	}

	// Version value
	version.Value = ctx.CurrentToken.Value
	ctx.CurrentToken = ctx.PeekToken
	ctx.PeekToken = ctx.Lexer.NextToken()

	return version, nil
}

func (g *PMFGrammar) parseAuthor(ctx *ParserContext) (*AuthorNode, error) {
	g.expectKeyword(ctx, "author")

	author := &AuthorNode{
		Span: ctx.CurrentToken.Span,
	}

	// Author value
	author.Name = ctx.CurrentToken.Value
	ctx.CurrentToken = ctx.PeekToken
	ctx.PeekToken = ctx.Lexer.NextToken()

	return author, nil
}

func (g *PMFGrammar) parseDescription(ctx *ParserContext) (*DescriptionNode, error) {
	g.expectKeyword(ctx, "description")

	description := &DescriptionNode{
		Span: ctx.CurrentToken.Span,
	}

	// Description text
	var text string
	for !g.check(ctx, lexer.TokenEOF) && !g.isSectionStart(ctx) {
		text += ctx.CurrentToken.Value + " "
		ctx.CurrentToken = ctx.PeekToken
		ctx.PeekToken = ctx.Lexer.NextToken()
	}

	description.Text = text
	return description, nil
}

func (g *PMFGrammar) parseInput(ctx *ParserContext) (*InputOutputNode, error) {
	g.expectKeyword(ctx, "input")

	io := &InputOutputNode{
		Span:    ctx.CurrentToken.Span,
		IsInput: true,
	}

	// Input name
	io.Name = ctx.CurrentToken.Value
	ctx.CurrentToken = ctx.PeekToken
	ctx.PeekToken = ctx.Lexer.NextToken()

	// Parse type
	if g.checkKeyword(ctx, "type") {
		g.expectKeyword(ctx, "type")
		io.Type = ctx.CurrentToken.Value
		ctx.CurrentToken = ctx.PeekToken
		ctx.PeekToken = ctx.Lexer.NextToken()
	}

	// Parse required/optional
	if g.checkKeyword(ctx, "required") {
		io.Required = true
		ctx.CurrentToken = ctx.PeekToken
		ctx.PeekToken = ctx.Lexer.NextToken()
	} else if g.checkKeyword(ctx, "optional") {
		io.Required = false
		ctx.CurrentToken = ctx.PeekToken
		ctx.PeekToken = ctx.Lexer.NextToken()
	}

	return io, nil
}

func (g *PMFGrammar) parseOutput(ctx *ParserContext) (*InputOutputNode, error) {
	g.expectKeyword(ctx, "output")

	io := &InputOutputNode{
		Span:    ctx.CurrentToken.Span,
		IsInput: false,
	}

	// Output name
	io.Name = ctx.CurrentToken.Value
	ctx.CurrentToken = ctx.PeekToken
	ctx.PeekToken = ctx.Lexer.NextToken()

	// Parse type
	if g.checkKeyword(ctx, "type") {
		g.expectKeyword(ctx, "type")
		io.Type = ctx.CurrentToken.Value
		ctx.CurrentToken = ctx.PeekToken
		ctx.PeekToken = ctx.Lexer.NextToken()
	}

	return io, nil
}

func (g *PMFGrammar) parseCapability(ctx *ParserContext) (*CapabilityNode, error) {
	g.expectKeyword(ctx, "capability")

	capability := &CapabilityNode{
		Span: ctx.CurrentToken.Span,
	}

	// Capability name
	capability.Name = ctx.CurrentToken.Value
	ctx.CurrentToken = ctx.PeekToken
	ctx.PeekToken = ctx.Lexer.NextToken()

	return capability, nil
}

func (g *PMFGrammar) isSectionStart(ctx *ParserContext) bool {
	return g.checkKeyword(ctx, "version") || g.checkKeyword(ctx, "author") ||
		g.checkKeyword(ctx, "description") || g.checkKeyword(ctx, "input") ||
		g.checkKeyword(ctx, "output") || g.checkKeyword(ctx, "capability")
}

// Node implementations for PMF

type ProcessorManifestNode struct {
	Name         string
	Version      *VersionNode
	Author       *AuthorNode
	Description  *DescriptionNode
	Inputs       []*InputOutputNode
	Outputs      []*InputOutputNode
	Capabilities []*CapabilityNode
	Span         lexer.Span
}

func (n *ProcessorManifestNode) Type() NodeType { return NodeType("ProcessorManifest") }
func (n *ProcessorManifestNode) Span() lexer.Span { return n.Span }
func (n *ProcessorManifestNode) Children() []Node {
	var children []Node
	if n.Version != nil {
		children = append(children, n.Version)
	}
	if n.Author != nil {
		children = append(children, n.Author)
	}
	if n.Description != nil {
		children = append(children, n.Description)
	}
	for _, io := range n.Inputs {
		children = append(children, io)
	}
	for _, io := range n.Outputs {
		children = append(children, io)
	}
	for _, cap := range n.Capabilities {
		children = append(children, cap)
	}
	return children
}

type VersionNode struct {
	Value string
	Span  lexer.Span
}

func (n *VersionNode) Type() NodeType { return NodeType("Version") }
func (n *VersionNode) Span() lexer.Span { return n.Span }
func (n *VersionNode) Children() []Node { return nil }

type AuthorNode struct {
	Name string
	Span lexer.Span
}

func (n *AuthorNode) Type() NodeType { return NodeType("Author") }
func (n *AuthorNode) Span() lexer.Span { return n.Span }
func (n *AuthorNode) Children() []Node { return nil }

type DescriptionNode struct {
	Text string
	Span lexer.Span
}

func (n *DescriptionNode) Type() NodeType { return NodeType("Description") }
func (n *DescriptionNode) Span() lexer.Span { return n.Span }
func (n *DescriptionNode) Children() []Node { return nil }

type InputOutputNode struct {
	Name     string
	Type     string
	Required bool
	IsInput  bool
	Span     lexer.Span
}

func (n *InputOutputNode) Type() NodeType {
	if n.IsInput {
		return NodeType("Input")
	}
	return NodeType("Output")
}
func (n *InputOutputNode) Span() lexer.Span { return n.Span }
func (n *InputOutputNode) Children() []Node { return nil }

type CapabilityNode struct {
	Name string
	Span lexer.Span
}

func (n *CapabilityNode) Type() NodeType { return NodeType("Capability") }
func (n *CapabilityNode) Span() lexer.Span { return n.Span }
func (n *CapabilityNode) Children() []Node { return nil }
