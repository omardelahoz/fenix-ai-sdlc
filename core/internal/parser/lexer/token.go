package lexer

import "fmt"

// TokenType represents the type of a token.
type TokenType int

const (
	// Special tokens
	TokenEOF TokenType = iota
	TokenError
	TokenComment
	TokenWhitespace

	// Literals
	TokenString
	TokenNumber
	TokenBoolean
	TokenNull

	// Identifiers and keywords
	TokenIdentifier
	TokenKeyword

	// Operators
	TokenEquals
	TokenColon
	TokenArrow
	TokenPipe
	TokenPlus
	TokenMinus
	TokenAsterisk
	TokenSlash
	TokenPercent
	TokenLessThan
	TokenGreaterThan
	TokenLessEqual
	TokenGreaterEqual
	TokenEqualsEquals
	TokenNotEquals
	TokenAnd
	TokenOr
	TokenNot
	TokenDot
	TokenComma
	TokenSemicolon

	// Delimiters
	TokenLeftParen
	TokenRightParen
	TokenLeftBrace
	TokenRightBrace
	TokenLeftBracket
	TokenRightBracket
	TokenLeftAngle
	TokenRightAngle

	// FDL-specific keywords
	TokenProduct
	TokenVision
	TokenUsers
	TokenConstraints
	TokenFeature
	TokenStory
	TokenRequirement
	TokenAcceptance
	TokenArchitecture
	TokenImplementation
	TokenRelease

	// WDL-specific keywords
	TokenWorkflow
	TokenStage
	TokenProcessor
	TokenPolicy
	TokenDepends
	TokenFanout
	TokenStrategy
	TokenOn
	TokenPipeline
	TokenGate
	TokenTimeout
	TokenApprovers

	// PMF-specific keywords
	TokenProcessorManifest
	TokenInput
	TokenOutput
	TokenCapability
	TokenVersion
	TokenAuthor
	TokenDescription

	// At-sign attributes
	TokenAtSign
)

// String returns the string representation of the token type.
func (t TokenType) String() string {
	switch t {
	case TokenEOF:
		return "EOF"
	case TokenError:
		return "ERROR"
	case TokenComment:
		return "COMMENT"
	case TokenWhitespace:
		return "WHITESPACE"
	case TokenString:
		return "STRING"
	case TokenNumber:
		return "NUMBER"
	case TokenBoolean:
		return "BOOLEAN"
	case TokenNull:
		return "NULL"
	case TokenIdentifier:
		return "IDENTIFIER"
	case TokenKeyword:
		return "KEYWORD"
	case TokenEquals:
		return "="
	case TokenColon:
		return ":"
	case TokenArrow:
		return "->"
	case TokenPipe:
		return "|"
	case TokenPlus:
		return "+"
	case TokenMinus:
		return "-"
	case TokenAsterisk:
		return "*"
	case TokenSlash:
		return "/"
	case TokenPercent:
		return "%"
	case TokenLessThan:
		return "<"
	case TokenGreaterThan:
		return ">"
	case TokenLessEqual:
		return "<="
	case TokenGreaterEqual:
		return ">="
	case TokenEqualsEquals:
		return "=="
	case TokenNotEquals:
		return "!="
	case TokenAnd:
		return "&&"
	case TokenOr:
		return "||"
	case TokenNot:
		return "!"
	case TokenDot:
		return "."
	case TokenComma:
		return ","
	case TokenSemicolon:
		return ";"
	case TokenLeftParen:
		return "("
	case TokenRightParen:
		return ")"
	case TokenLeftBrace:
		return "{"
	case TokenRightBrace:
		return "}"
	case TokenLeftBracket:
		return "["
	case TokenRightBracket:
		return "]"
	case TokenLeftAngle:
		return "<"
	case TokenRightAngle:
	return ">"
	case TokenAtSign:
		return "@"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", t)
	}
}

// IsKeyword returns true if the token type is a keyword.
func (t TokenType) IsKeyword() bool {
	return t == TokenKeyword ||
		t == TokenProduct || t == TokenVision || t == TokenUsers ||
		t == TokenConstraints || t == TokenFeature || t == TokenStory ||
		t == TokenRequirement || t == TokenAcceptance || t == TokenArchitecture ||
		t == TokenImplementation || t == TokenRelease ||
		t == TokenWorkflow || t == TokenStage || t == TokenProcessor ||
		t == TokenPolicy || t == TokenDepends || t == TokenFanout ||
		t == TokenStrategy || t == TokenOn || t == TokenPipeline ||
		t == TokenGate || t == TokenTimeout || t == TokenApprovers ||
		t == TokenProcessorManifest || t == TokenInput || t == TokenOutput ||
		t == TokenCapability || t == TokenVersion || t == TokenAuthor ||
		t == TokenDescription
}

// IsLiteral returns true if the token type is a literal.
func (t TokenType) IsLiteral() bool {
	return t == TokenString || t == TokenNumber || t == TokenBoolean || t == TokenNull
}

// IsOperator returns true if the token type is an operator.
func (t TokenType) IsOperator() bool {
	return t == TokenEquals || t == TokenColon || t == TokenArrow ||
		t == TokenPipe || t == TokenPlus || t == TokenMinus ||
		t == TokenAsterisk || t == TokenSlash || t == TokenPercent ||
		t == TokenLessThan || t == TokenGreaterThan || t == TokenLessEqual ||
		t == TokenGreaterEqual || t == TokenEqualsEquals || t == TokenNotEquals ||
		t == TokenAnd || t == TokenOr || t == TokenNot || t == TokenDot
}

// IsDelimiter returns true if the token type is a delimiter.
func (t TokenType) IsDelimiter() bool {
	return t == TokenLeftParen || t == TokenRightParen ||
		t == TokenLeftBrace || t == TokenRightBrace ||
		t == TokenLeftBracket || t == TokenRightBracket ||
		t == TokenComma || t == TokenSemicolon
}
