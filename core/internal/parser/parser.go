package parser

import (
	"fmt"

	"github.com/omardelahoz/fenix-ai-sdlc/core/internal/parser/grammar"
	"github.com/omardelahoz/fenix-ai-sdlc/core/internal/parser/lexer"
)

// Parser is the generic parsing engine that uses a Grammar to parse source text.
// It implements an LL(1) recursive descent parser with error recovery.
type Parser struct {
	grammar grammar.Grammar
	ctx     *grammar.ParserContext
}

// NewParser creates a new parser with the given grammar.
func NewParser(grammar grammar.Grammar, source, fileName string) *Parser {
	lex := lexer.NewLexer(source, fileName)

	ctx := &grammar.ParserContext{
		Lexer: lex,
	}

	// Initialize tokens
	ctx.CurrentToken = lex.NextToken()
	ctx.PeekToken = lex.NextToken()

	return &Parser{
		grammar: grammar,
		ctx:     ctx,
	}
}

// Parse parses the source text using the grammar.
func (p *Parser) Parse() (grammar.Node, error) {
	node, err := p.grammar.ParseRoot(p.ctx)
	if err != nil {
		return nil, err
	}

	if len(p.ctx.Errors) > 0 {
		return node, &ParseError{
			Errors: p.ctx.Errors,
		}
	}

	return node, nil
}

// advance advances to the next token.
func (p *Parser) advance() {
	p.ctx.CurrentToken = p.ctx.PeekToken
	p.ctx.PeekToken = p.ctx.Lexer.NextToken()
}

// expect advances if the current token matches the expected type, otherwise adds an error.
func (p *Parser) expect(tokenType lexer.TokenType) bool {
	if p.ctx.CurrentToken.Type == tokenType {
		p.advance()
		return true
	}

	p.error(fmt.Sprintf("expected %s, got %s", tokenType, p.ctx.CurrentToken.Type))
	return false
}

// error adds a parse error.
func (p *Parser) error(message string) {
	p.ctx.Errors = append(p.ctx.Errors, grammar.ParseError{
		Message: message,
		Span:    p.ctx.CurrentToken.Span,
	})
}

// expectKeyword advances if the current token is the expected keyword.
func (p *Parser) expectKeyword(keyword string) bool {
	if p.ctx.CurrentToken.Type == lexer.TokenKeyword && p.ctx.CurrentToken.Value == keyword {
		p.advance()
		return true
	}

	p.error(fmt.Sprintf("expected keyword '%s', got '%s'", keyword, p.ctx.CurrentToken.Value))
	return false
}

// check returns true if the current token matches the expected type.
func (p *Parser) check(tokenType lexer.TokenType) bool {
	return p.ctx.CurrentToken.Type == tokenType
}

// checkKeyword returns true if the current token is the expected keyword.
func (p *Parser) checkKeyword(keyword string) bool {
	return p.ctx.CurrentToken.Type == lexer.TokenKeyword && p.ctx.CurrentToken.Value == keyword
}

// match advances if the current token matches any of the given types.
func (p *Parser) match(types ...lexer.TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

// synchronize attempts to recover from an error by synchronizing to a valid point.
func (p *Parser) synchronize() {
	p.advance()

	for !p.check(lexer.TokenEOF) {
		if p.ctx.CurrentToken.Type == lexer.TokenSemicolon {
			return
		}

		// Synchronize at block boundaries
		switch p.ctx.CurrentToken.Type {
		case lexer.TokenProduct, lexer.TokenWorkflow, lexer.TokenProcessorManifest,
			lexer.TokenFeature, lexer.TokenStage, lexer.TokenRelease:
			return
		}

		p.advance()
	}
}

// ParseError represents a parsing error with multiple sub-errors.
type ParseError struct {
	Errors []grammar.ParseError
}

func (e *ParseError) Error() string {
	if len(e.Errors) == 0 {
		return "parse error"
	}
	return fmt.Sprintf("parse error: %d error(s)", len(e.Errors))
}
