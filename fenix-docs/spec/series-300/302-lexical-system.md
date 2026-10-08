# SPEC-302: Lexical System

## 1. Design Goals
The Fénix Lexical System is the first stage of the compilation pipeline (SPEC-300). Its sole responsibility is to consume raw UTF-8 source text and produce a flat stream of `Token`s, enriching them with `Trivia` (whitespace, comments). 

Key design goals:
- **Language Agnosticism:** The Lexer does not know FDL, WDL, or PMF keywords. It emits generic `TokenIdentifier`. Keyword resolution is exclusively the Parser's job.
- **Zero Allocations & GC Immunity:** The Lexer operates directly over `[]byte` and allocates tokens into an Arena (SPEC-301). It avoids expanding the source into a `[]rune` array.
- **Full Fidelity:** 100% of the source text, including every space and comment, is preserved in the Trivia model.
- **O(N) Complexity:** A pure State Machine without backtracking, lookahead limits, or Regular Expressions.

## 2. Lexical Architecture
The Lexical System operates in a single pass.
```text
SourceFile ([]byte)
       │
       ▼
 ┌───────────┐
 │   Lexer   │ ─► Consumes bytes, emits tokens and diagnostics.
 └───────────┘
       │
       ▼
   LexResult (Tokens, Diagnostics, EndState)
```
*Note: Internally, the Lexer may use helpers like `TextWindow`, `Scanner`, and `TokenBuilder`, but architecturally it is a single monolithic component.*

## 3. Source Buffer ([]byte)
**Architectural Rule:** The lexer operates directly over UTF-8 encoded `[]byte`. Unicode code points are decoded lazily *only* when required for identifier classification or escape processing. The original source buffer is never expanded into `[]rune`. 
This minimizes memory footprint, maximizes CPU cache locality, and is critical for sub-millisecond incremental lexing of massive monorepos.

## 4. State Machine
The Scanner is a simple, deterministic Finite State Machine (FSM). It evaluates the current byte and transitions states.

```text
Default
  │
  ├── Identifier   (a-z, A-Z, _, Unicode letters)
  ├── Number       (0-9)
  ├── Operator     (=, >=, <=, *, @)
  ├── Comment      (//)
  ├── DocComment   (///)
  ├── Whitespace   ( , \t, \r, \n)
  └── String       (" or """)
        │
        ├── Escape         (\", \n, \uXXXX)
        ├── Interpolation  (${)
        └── End            (" or """)
```
There are no nested states outside of String interpolation transitions.

## 5. UTF-8 Handling
While the main loop iterates over bytes, Fénix fully supports Unicode identifiers and strings.
- **Fast Path:** If `byte < 128` (ASCII), it's processed in `O(1)`.
- **Slow Path:** If `byte >= 128`, the Lexer reads a full `rune` via `utf8.DecodeRune` to validate if it is a valid identifier character or whitespace.

## 6. Trivia Model
Trivia represents parts of the source text that are insignificant to the compiler's semantic meaning but essential for the Formatter and IDE (LSP). They are classified into strict categories (Whitespace, Newline, Comment, DocComment).

- **Leading Trivia:** Any whitespace, newline, or comment that appears *before* a significant token on the same line, or on preceding lines.
- **Trailing Trivia:** Any whitespace or comment that appears *after* a significant token, up to the end of the current line.

**Example:**
```fdl
// A comment      <-- Leading Trivia of 'entity'
entity            <-- Token
  Customer        <-- Leading Trivia (spaces) + Token
  {               <-- Token + Trailing Trivia (spaces)
```

## 7. Token Model
The Lexer does not emit `TokenEntity` or `TokenService`. It emits:
- `TokenIdentifier` (including `true` and `false`, which are resolved semantically)
- `TokenStringLiteral`
- `TokenIntLiteral` / `TokenFloatLiteral`
- `TokenSymbol` (e.g., `{`, `}`, `.`)
- `TokenOperator` (e.g., `=`)
- Synthetic tokens like `TokenMissing` for error recovery.

## 8. String Literals (Single & Multiline)
Fénix supports two types of strings natively:
1. **Single-line Strings:** Delimited by `"`. Cannot contain raw unescaped newlines.
2. **Multiline Strings:** Delimited by `"""`. Preserves all internal formatting, newlines, and whitespace. Highly useful for documentation, SQL injection, and prompts.

### Escapes
Only standard escapes are supported to keep the language surface small:
`\\`, `\"`, `\n`, `\r`, `\t`, `\uXXXX`.

## 9. String Interpolation (${expression})
Fénix supports string interpolation from V1 to future-proof the language for dynamic configurations, routes, and prompts.
- **Syntax:** Strictly `${expression}`. Bash-style `$name` or template-style `{{name}}` are forbidden to keep parsing unambiguous.
- **Nesting:** The Lexer maintains an internal `{` `}` nesting counter while in the Interpolation state. This allows complex inner expressions like `${foo({ })}` without prematurely terminating the interpolation.
- **Lexer Responsibility:** The Lexer does **not** evaluate expressions. It emits boundaries.
```fdl
"Hello ${customer.name}"
```
Produces:
1. `TokenStringStart`
2. `TokenStringText` ("Hello ")
3. `TokenInterpolationStart` (`${`)
4. `TokenIdentifier` ("customer")
5. `TokenSymbol` (`.`)
6. `TokenIdentifier` ("name")
7. `TokenInterpolationEnd` (`}`)
8. `TokenStringEnd`

The Parser controls when interpolation ends and reuses the exact same `ExpressionParser` used elsewhere in the language. The Lexer simply yields tokens.

## 10. Error Recovery
The Lexer never crashes. If it encounters illegal bytes (e.g., `\0`), an unclosed string, or an invalid escape sequence:
1. It emits a generic `TokenIdentifier` or `TokenStringText`.
2. It pushes a `Diagnostic` (e.g., `LEX1001: Unterminated string`) into the `LexResult.Diagnostics` stream, rather than inflating the `Token` struct.
This ensures the Parser always receives a continuous stream and can build a complete (albeit marked as erroneous) CST.

## 11. Incremental Lexing
By keeping the Lexer completely stateless and agnostic, a change in a single line of a massive file allows the system to re-lex **only** the modified chunk of bytes. The Lexer can be seeded with an initial state (e.g., `InString`) from the previous syntax tree's end-of-line state, and it returns an `EndState` to link with the next unchanged block.

## 12. Performance Considerations
- **No String Allocation:** `Token` structs do not contain a `Lexeme` string. The token uses its `Span` to reference the original `[]byte` slice when the text is actually needed (e.g., `token.Text(source)`), saving massive GC overhead.
- **Tight Loop:** The `Lexer.Next()` function is a massive `switch` statement optimized by the Go compiler.
- **Zero Regex:** The `regexp` package is strictly forbidden in the Lexer path.
