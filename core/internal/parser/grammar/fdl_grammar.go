package grammar

import (
	"github.com/omardelahoz/fenix-ai-sdlc/core/internal/parser/lexer"
)

// FDLGrammar implements the Grammar interface for FDL (Fénix Definition Language).
type FDLGrammar struct{}

// NewFDLGrammar creates a new FDL grammar.
func NewFDLGrammar() *FDLGrammar {
	return &FDLGrammar{}
}

// Name returns the name of the grammar.
func (g *FDLGrammar) Name() string {
	return "FDL"
}

// ParseRoot parses the root of an FDL document.
func (g *FDLGrammar) ParseRoot(ctx *ParserContext) (Node, error) {
	// FDL document starts with "product <name>"
	if !g.expectKeyword(ctx, "product") {
		return nil, ctx.Errors[0]
	}

	// Product name
	if !g.expect(ctx, lexer.TokenIdentifier) {
		return nil, ctx.Errors[0]
	}

	// Create document node
	doc := &DocumentNode{
		ProductName: ctx.CurrentToken.Value,
		Span:        ctx.CurrentToken.Span,
	}

	// Parse sections
	for !g.check(ctx, lexer.TokenEOF) {
		switch {
		case g.checkKeyword(ctx, "vision"):
			vision, err := g.parseVision(ctx)
			if err != nil {
				return nil, err
			}
			doc.Vision = vision
		case g.checkKeyword(ctx, "users"):
			users, err := g.parseUsers(ctx)
			if err != nil {
				return nil, err
			}
			doc.Users = users
		case g.checkKeyword(ctx, "constraints"):
			constraints, err := g.parseConstraints(ctx)
			if err != nil {
				return nil, err
			}
			doc.Constraints = constraints
		case g.checkKeyword(ctx, "feature"):
			feature, err := g.parseFeature(ctx)
			if err != nil {
				return nil, err
			}
			doc.Features = append(doc.Features, feature)
		case g.checkKeyword(ctx, "implementation"):
			impl, err := g.parseImplementation(ctx)
			if err != nil {
				return nil, err
			}
			doc.Implementation = impl
		case g.checkKeyword(ctx, "release"):
			release, err := g.parseRelease(ctx)
			if err != nil {
				return nil, err
			}
			doc.Release = release
		default:
			// Skip unknown tokens
			ctx.CurrentToken = ctx.PeekToken
			ctx.PeekToken = ctx.Lexer.NextToken()
		}
	}

	return doc, nil
}

// IsKeyword returns true if the identifier is an FDL keyword.
func (g *FDLGrammar) IsKeyword(ident string) bool {
	switch ident {
	case "product", "vision", "users", "constraints",
		"feature", "story", "requirement", "acceptance",
		"architecture", "implementation", "release":
		return true
	default:
		return false
	}
}

// GetKeywordType returns the token type for a keyword.
func (g *FDLGrammar) GetKeywordType(ident string) lexer.TokenType {
	switch ident {
	case "product":
		return lexer.TokenProduct
	case "vision":
		return lexer.TokenVision
	case "users":
		return lexer.TokenUsers
	case "constraints":
		return lexer.TokenConstraints
	case "feature":
		return lexer.TokenFeature
	case "story":
		return lexer.TokenStory
	case "requirement":
		return lexer.TokenRequirement
	case "acceptance":
		return lexer.TokenAcceptance
	case "architecture":
		return lexer.TokenArchitecture
	case "implementation":
		return lexer.TokenImplementation
	case "release":
		return lexer.TokenRelease
	default:
		return lexer.TokenIdentifier
	}
}

// Helper methods for parsing FDL structures

func (g *FDLGrammar) expectKeyword(ctx *ParserContext, keyword string) bool {
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

func (g *FDLGrammar) expect(ctx *ParserContext, tokenType lexer.TokenType) bool {
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

func (g *FDLGrammar) check(ctx *ParserContext, tokenType lexer.TokenType) bool {
	return ctx.CurrentToken.Type == tokenType
}

func (g *FDLGrammar) checkKeyword(ctx *ParserContext, keyword string) bool {
	return ctx.CurrentToken.Type == lexer.TokenKeyword && ctx.CurrentToken.Value == keyword
}

func (g *FDLGrammar) parseVision(ctx *ParserContext) (*VisionNode, error) {
	g.expectKeyword(ctx, "vision")
	
	vision := &VisionNode{
		Span: ctx.CurrentToken.Span,
	}
	
	// Read until next keyword or EOF
	var text string
	for !g.check(ctx, lexer.TokenEOF) && !g.isSectionStart(ctx) {
		text += ctx.CurrentToken.Value + " "
		ctx.CurrentToken = ctx.PeekToken
		ctx.PeekToken = ctx.Lexer.NextToken()
	}
	
	vision.Text = text
	return vision, nil
}

func (g *FDLGrammar) parseUsers(ctx *ParserContext) (*UsersNode, error) {
	g.expectKeyword(ctx, "users")
	
	users := &UsersNode{
		Span: ctx.CurrentToken.Span,
	}
	
	// Parse user definitions
	for !g.check(ctx, lexer.TokenEOF) && !g.isSectionStart(ctx) {
		userName := ctx.CurrentToken.Value
		users.UserNames = append(users.UserNames, userName)
		ctx.CurrentToken = ctx.PeekToken
		ctx.PeekToken = ctx.Lexer.NextToken()
	}
	
	return users, nil
}

func (g *FDLGrammar) parseConstraints(ctx *ParserContext) (*ConstraintsNode, error) {
	g.expectKeyword(ctx, "constraints")
	
	constraints := &ConstraintsNode{
		Span: ctx.CurrentToken.Span,
	}
	
	// Parse constraint definitions
	for !g.check(ctx, lexer.TokenEOF) && !g.isSectionStart(ctx) {
		constraintName := ctx.CurrentToken.Value
		constraints.Names = append(constraints.Names, constraintName)
		ctx.CurrentToken = ctx.PeekToken
		ctx.PeekToken = ctx.Lexer.NextToken()
	}
	
	return constraints, nil
}

func (g *FDLGrammar) parseFeature(ctx *ParserContext) (*FeatureNode, error) {
	g.expectKeyword(ctx, "feature")
	
	feature := &FeatureNode{
		Span: ctx.CurrentToken.Span,
	}
	
	// Feature name
	feature.Name = ctx.CurrentToken.Value
	ctx.CurrentToken = ctx.PeekToken
	ctx.PeekToken = ctx.Lexer.NextToken()
	
	// Parse stories
	for g.checkKeyword(ctx, "story") {
		story, err := g.parseStory(ctx)
		if err != nil {
			return nil, err
		}
		feature.Stories = append(feature.Stories, story)
	}
	
	return feature, nil
}

func (g *FDLGrammar) parseStory(ctx *ParserContext) (*StoryNode, error) {
	g.expectKeyword(ctx, "story")
	
	story := &StoryNode{
		Span: ctx.CurrentToken.Span,
	}
	
	// Story name
	story.Name = ctx.CurrentToken.Value
	ctx.CurrentToken = ctx.PeekToken
	ctx.PeekToken = ctx.Lexer.NextToken()
	
	// Parse story content
	for !g.check(ctx, lexer.TokenEOF) && !g.isSectionStart(ctx) && !g.checkKeyword(ctx, "story") {
		ctx.CurrentToken = ctx.PeekToken
		ctx.PeekToken = ctx.Lexer.NextToken()
	}
	
	return story, nil
}

func (g *FDLGrammar) parseImplementation(ctx *ParserContext) (*ImplementationNode, error) {
	g.expectKeyword(ctx, "implementation")
	
	impl := &ImplementationNode{
		Span: ctx.CurrentToken.Span,
	}
	
	// Parse implementation content
	for !g.check(ctx, lexer.TokenEOF) && !g.isSectionStart(ctx) {
		ctx.CurrentToken = ctx.PeekToken
		ctx.PeekToken = ctx.Lexer.NextToken()
	}
	
	return impl, nil
}

func (g *FDLGrammar) parseRelease(ctx *ParserContext) (*ReleaseNode, error) {
	g.expectKeyword(ctx, "release")
	
	release := &ReleaseNode{
		Span: ctx.CurrentToken.Span,
	}
	
	// Release version
	release.Version = ctx.CurrentToken.Value
	ctx.CurrentToken = ctx.PeekToken
	ctx.PeekToken = ctx.Lexer.NextToken()
	
	return release, nil
}

func (g *FDLGrammar) isSectionStart(ctx *ParserContext) bool {
	return g.checkKeyword(ctx, "vision") || g.checkKeyword(ctx, "users") ||
		g.checkKeyword(ctx, "constraints") || g.checkKeyword(ctx, "feature") ||
		g.checkKeyword(ctx, "implementation") || g.checkKeyword(ctx, "release")
}

// Node implementations for FDL

type DocumentNode struct {
	ProductName    string
	Vision         *VisionNode
	Users          *UsersNode
	Constraints    *ConstraintsNode
	Features       []*FeatureNode
	Implementation *ImplementationNode
	Release        *ReleaseNode
	Span           lexer.Span
}

func (n *DocumentNode) Type() NodeType { return NodeTypeDocument }
func (n *DocumentNode) Span() lexer.Span { return n.Span }
func (n *DocumentNode) Children() []Node {
	var children []Node
	if n.Vision != nil {
		children = append(children, n.Vision)
	}
	if n.Users != nil {
		children = append(children, n.Users)
	}
	if n.Constraints != nil {
		children = append(children, n.Constraints)
	}
	for _, f := range n.Features {
		children = append(children, f)
	}
	if n.Implementation != nil {
		children = append(children, n.Implementation)
	}
	if n.Release != nil {
		children = append(children, n.Release)
	}
	return children
}

type VisionNode struct {
	Text string
	Span lexer.Span
}

func (n *VisionNode) Type() NodeType { return NodeTypeVision }
func (n *VisionNode) Span() lexer.Span { return n.Span }
func (n *VisionNode) Children() []Node { return nil }

type UsersNode struct {
	UserNames []string
	Span      lexer.Span
}

func (n *UsersNode) Type() NodeType { return NodeTypeUsers }
func (n *UsersNode) Span() lexer.Span { return n.Span }
func (n *UsersNode) Children() []Node { return nil }

type ConstraintsNode struct {
	Names []string
	Span  lexer.Span
}

func (n *ConstraintsNode) Type() NodeType { return NodeTypeConstraints }
func (n *ConstraintsNode) Span() lexer.Span { return n.Span }
func (n *ConstraintsNode) Children() []Node { return nil }

type FeatureNode struct {
	Name    string
	Stories []*StoryNode
	Span    lexer.Span
}

func (n *FeatureNode) Type() NodeType { return NodeTypeFeature }
func (n *FeatureNode) Span() lexer.Span { return n.Span }
func (n *FeatureNode) Children() []Node {
	var children []Node
	for _, s := range n.Stories {
		children = append(children, s)
	}
	return children
}

type StoryNode struct {
	Name string
	Span lexer.Span
}

func (n *StoryNode) Type() NodeType { return NodeTypeStory }
func (n *StoryNode) Span() lexer.Span { return n.Span }
func (n *StoryNode) Children() []Node { return nil }

type ImplementationNode struct {
	Span lexer.Span
}

func (n *ImplementationNode) Type() NodeType { return NodeTypeImplementation }
func (n *ImplementationNode) Span() lexer.Span { return n.Span }
func (n *ImplementationNode) Children() []Node { return nil }

type ReleaseNode struct {
	Version string
	Span    lexer.Span
}

func (n *ReleaseNode) Type() NodeType { return NodeTypeRelease }
func (n *ReleaseNode) Span() lexer.Span { return n.Span }
func (n *ReleaseNode) Children() []Node { return nil }
