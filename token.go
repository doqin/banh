package main

type TokenType string

const (
	TokenEOF       TokenType = "EOF"
	TokenKeyword   TokenType = "KEYWORD"
	TokenIdent     TokenType = "IDENTIFIER"
	TokenNumber    TokenType = "NUMBER"
	TokenString    TokenType = "STRING"
	TokenOperator  TokenType = "OPERATOR"
	TokenLParen    TokenType = "LPAREN"
	TokenRParen    TokenType = "RPAREN"
	TokenLBrack    TokenType = "LBRACK"
	TokenRBrack    TokenType = "RBRACK"
	TokenDotDot    TokenType = "DOTDOT"
	TokenComma     TokenType = "COMMA"
	TokenSemiColon TokenType = "SEMICOLON"
	TokenNewLine   TokenType = "NEWLINE"
	TokenPrimitive TokenType = "PRIMITIVE"
	TokenContainer TokenType = "CONTAINER"
	TokenUnknown   TokenType = "UNKNOWN"
)

type Token struct {
	Type   TokenType
	Lexeme string
	Line   int
	Column int
}

var Operators = map[string]string{
	"+": SymbolPlus,
	"-": SymbolMinus,
	"*": SymbolAsterisk,
	"/": SymbolSlash,
	"%": SymbolModulo,
	"!": SymbolBang,
	"=": SymbolEqual,
	"<": SymbolLess,
	">": SymbolGreater,
}

var Symbols = map[string]TokenType{
	"(":  TokenLParen,
	")":  TokenRParen,
	"[":  TokenLBrack,
	"]":  TokenRBrack,
	"..": TokenDotDot,
	";":  TokenSemiColon,
	",":  TokenComma,
}
