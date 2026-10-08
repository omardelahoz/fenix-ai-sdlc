package lexer

import (
	"unicode"
	"unicode/utf8"
)

// Position represents a position in the source text.
type Position struct {
	Line   int // 1-based
	Column int // 1-based (in bytes)
	Offset int // 0-based byte offset
}

// Span represents a range in the source text.
type Span struct {
	Start Position
	End   Position
}

// Trivia represents whitespace or comments between tokens.
type Trivia struct {
	Kind TriviaKind
	Text string
	Span Span
}

// TriviaKind represents the type of trivia.
type TriviaKind int

const (
	TriviaWhitespace TriviaKind = iota
	TriviaComment
	TriviaNewline
)

// Token represents a lexical token with optional trivia.
type Token struct {
	Type     TokenType
	Value    string // The actual text of the token
	Span     Span
	Leading  []Trivia // Trivia before the token
	Trailing []Trivia // Trivia after the token
}

// Lexer tokenizes source text into tokens.
type Lexer struct {
	source   string
	pos      Position
	ch       rune
	atEOF    bool
	fileName string
}

// NewLexer creates a new lexer for the given source text.
func NewLexer(source, fileName string) *Lexer {
	l := &Lexer{
		source:   source,
		pos:      Position{Line: 1, Column: 1, Offset: 0},
		fileName: fileName,
	}
	l.readChar()
	return l
}

// readChar reads the next character from the source.
func (l *Lexer) readChar() {
	if l.pos.Offset >= len(l.source) {
		l.atEOF = true
		l.ch = 0
		return
	}

	l.ch, _ = l.at(l.pos.Offset)
	l.pos.Offset++
}

// at returns the rune at the given offset.
func (l *Lexer) at(offset int) (rune, int) {
	if offset >= len(l.source) {
		return 0, 0
	}
	return utf8.DecodeRuneInString(l.source[offset:])
}

// peek returns the next character without consuming it.
func (l *Lexer) peek() rune {
	if l.atEOF {
		return 0
	}
	ch, _ := l.at(l.pos.Offset)
	return ch
}

// peekAhead peeks n characters ahead.
func (l *Lexer) peekAhead(n int) rune {
	if l.pos.Offset+n >= len(l.source) {
		return 0
	}
	ch, _ := l.at(l.pos.Offset + n)
	return ch
}

// NextToken returns the next token from the source.
func (l *Lexer) NextToken() Token {
	// Skip whitespace but collect it as trivia
	l.skipWhitespace()

	startPos := l.pos

	var tok Token
	tok.Span.Start = startPos

	switch l.ch {
	case 0:
		tok.Type = TokenEOF
	case '#':
		tok = l.readComment()
	case '"':
		tok = l.readString()
	case '\'':
		tok = l.readString()
	case '{':
		tok.Type = TokenLeftBrace
		tok.Value = "{"
		l.readChar()
	case '}':
		tok.Type = TokenRightBrace
		tok.Value = "}"
		l.readChar()
	case '(':
		tok.Type = TokenLeftParen
		tok.Value = "("
		l.readChar()
	case ')':
		tok.Type = TokenRightParen
		tok.Value = ")"
		l.readChar()
	case '[':
		tok.Type = TokenLeftBracket
		tok.Value = "["
		l.readChar()
	case ']':
		tok.Type = TokenRightBracket
		tok.Value = "]"
		l.readChar()
	case '=':
		if l.peek() == '=' {
			tok.Type = TokenEqualsEquals
			tok.Value = "=="
			l.readChar()
			l.readChar()
		} else {
			tok.Type = TokenEquals
			tok.Value = "="
			l.readChar()
		}
	case ':':
		tok.Type = TokenColon
		tok.Value = ":"
		l.readChar()
	case '-':
		if l.peek() == '>' {
			tok.Type = TokenArrow
			tok.Value = "->"
			l.readChar()
			l.readChar()
		} else {
			tok.Type = TokenMinus
			tok.Value = "-"
			l.readChar()
		}
	case '+':
		tok.Type = TokenPlus
		tok.Value = "+"
		l.readChar()
	case '*':
		tok.Type = TokenAsterisk
		tok.Value = "*"
		l.readChar()
	case '/':
		tok.Type = TokenSlash
		tok.Value = "/"
		l.readChar()
	case '%':
		tok.Type = TokenPercent
		tok.Value = "%"
		l.readChar()
	case '<':
		if l.peek() == '=' {
			tok.Type = TokenLessEqual
			tok.Value = "<="
			l.readChar()
			l.readChar()
		} else {
			tok.Type = TokenLessThan
			tok.Value = "<"
			l.readChar()
		}
	case '>':
		if l.peek() == '=' {
			tok.Type = TokenGreaterEqual
			tok.Value = ">="
			l.readChar()
			l.readChar()
		} else {
			tok.Type = TokenGreaterThan
			tok.Value = ">"
			l.readChar()
		}
	case '!':
		if l.peek() == '=' {
			tok.Type = TokenNotEquals
			tok.Value = "!="
			l.readChar()
			l.readChar()
		} else {
			tok.Type = TokenNot
			tok.Value = "!"
			l.readChar()
		}
	case '&':
		if l.peek() == '&' {
			tok.Type = TokenAnd
			tok.Value = "&&"
			l.readChar()
			l.readChar()
		} else {
			tok.Type = TokenError
			tok.Value = "&"
			l.readChar()
		}
	case '|':
		if l.peek() == '|' {
			tok.Type = TokenOr
			tok.Value = "||"
			l.readChar()
			l.readChar()
		} else {
			tok.Type = TokenPipe
			tok.Value = "|"
			l.readChar()
		}
	case '.':
		tok.Type = TokenDot
		tok.Value = "."
		l.readChar()
	case ',':
		tok.Type = TokenComma
		tok.Value = ","
		l.readChar()
	case ';':
		tok.Type = TokenSemicolon
		tok.Value = ";"
		l.readChar()
	case '@':
		tok.Type = TokenAtSign
		tok.Value = "@"
		l.readChar()
	default:
		if unicode.IsDigit(l.ch) {
			tok = l.readNumber()
		} else if unicode.IsLetter(l.ch) || l.ch == '_' {
			tok = l.readIdentifier()
		} else {
			tok.Type = TokenError
			tok.Value = string(l.ch)
			l.readChar()
		}
	}

	tok.Span.End = l.pos
	return tok
}

// skipWhitespace skips whitespace characters.
func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.ch) && !l.atEOF {
		if l.ch == '\n' {
			l.pos.Line++
			l.pos.Column = 1
		} else {
			l.pos.Column++
		}
		l.readChar()
	}
}

// readComment reads a comment line.
func (l *Lexer) readComment() Token {
	startPos := l.pos
	l.readChar() // Skip '#'

	var comment string
	for l.ch != '\n' && l.ch != 0 && !l.atEOF {
		comment += string(l.ch)
		l.readChar()
	}

	return Token{
		Type:  TokenComment,
		Value: comment,
		Span: Span{
			Start: startPos,
			End:   l.pos,
		},
	}
}

// readString reads a string literal.
func (l *Lexer) readString() Token {
	startPos := l.pos
	quote := l.ch
	l.readChar() // Skip opening quote

	var str string
	for l.ch != quote && l.ch != 0 && !l.atEOF {
		if l.ch == '\\' {
			// Handle escape sequences
			l.readChar()
			switch l.ch {
			case 'n':
				str += "\n"
			case 't':
				str += "\t"
			case 'r':
				str += "\r"
			case '\\':
				str += "\\"
			case '"':
				str += "\""
			case '\'':
				str += "'"
			default:
				str += string(l.ch)
			}
		} else {
			str += string(l.ch)
		}
		l.readChar()
	}

	l.readChar() // Skip closing quote

	return Token{
		Type:  TokenString,
		Value: str,
		Span: Span{
			Start: startPos,
			End:   l.pos,
		},
	}
}

// readNumber reads a number literal.
func (l *Lexer) readNumber() Token {
	startPos := l.pos
	var num string

	for unicode.IsDigit(l.ch) && !l.atEOF {
		num += string(l.ch)
		l.readChar()
	}

	// Handle decimal point
	if l.ch == '.' && unicode.IsDigit(l.peek()) {
		num += string(l.ch)
		l.readChar()
		for unicode.IsDigit(l.ch) && !l.atEOF {
			num += string(l.ch)
			l.readChar()
		}
	}

	return Token{
		Type:  TokenNumber,
		Value: num,
		Span: Span{
			Start: startPos,
			End:   l.pos,
		},
	}
}

// readIdentifier reads an identifier or keyword.
func (l *Lexer) readIdentifier() Token {
	startPos := l.pos
	var ident string

	for (unicode.IsLetter(l.ch) || unicode.IsDigit(l.ch) || l.ch == '_') && !l.atEOF {
		ident += string(l.ch)
		l.readChar()
	}

	// Check if it's a keyword
	tokenType := l.lookupKeyword(ident)
	if tokenType == TokenIdentifier {
		tokenType = TokenKeyword
	}

	return Token{
		Type:  tokenType,
		Value: ident,
		Span: Span{
			Start: startPos,
			End:   l.pos,
		},
	}
}

// lookupKeyword checks if an identifier is a keyword.
func (l *Lexer) lookupKeyword(ident string) TokenType {
	switch ident {
	case "product":
		return TokenProduct
	case "vision":
		return TokenVision
	case "users":
		return TokenUsers
	case "constraints":
		return TokenConstraints
	case "feature":
		return TokenFeature
	case "story":
		return TokenStory
	case "requirement":
		return TokenRequirement
	case "acceptance":
		return TokenAcceptance
	case "architecture":
		return TokenArchitecture
	case "implementation":
		return TokenImplementation
	case "release":
		return TokenRelease
	case "workflow":
		return TokenWorkflow
	case "stage":
		return TokenStage
	case "processor":
		return TokenProcessor
	case "policy":
		return TokenPolicy
	case "depends":
		return TokenDepends
	case "fanout":
		return TokenFanout
	case "strategy":
		return TokenStrategy
	case "on":
		return TokenOn
	case "pipeline":
		return TokenPipeline
	case "gate":
		return TokenGate
	case "timeout":
		return TokenTimeout
	case "approvers":
		return TokenApprovers
	case "true", "false":
		return TokenBoolean
	case "null":
		return TokenNull
	default:
		return TokenIdentifier
	}
}

// Tokenize tokenizes the entire source text.
func (l *Lexer) Tokenize() []Token {
	var tokens []Token

	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)

		if tok.Type == TokenEOF || tok.Type == TokenError {
			break
		}
	}

	return tokens
}
