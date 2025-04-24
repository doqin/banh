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
	TokenComma     TokenType = "COMMA"
	TokenSemiColon TokenType = "SEMICOLON"
	TokenNewLine   TokenType = "NEWLINE"
	TokenInType    TokenType = "INTYPE"
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
	"(": TokenLParen,
	")": TokenRParen,
	";": TokenSemiColon,
	",": TokenComma,
}
