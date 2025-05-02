package main

import (
	"unicode"

	"golang.org/x/text/unicode/norm"
)

type Lexer struct {
	input []rune
	pos   int
	line  int
	col   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{
		input: []rune(norm.NFC.String(input)),
		line:  1,
		col:   1,
	}
}

// Main function
func (l *Lexer) NextToken() Token {
	newLine := l.skipWhitespace()
	if newLine != nil {
		return *newLine
	}

	if l.pos >= len(l.input) {
		return Token{Type: TokenEOF, Lexeme: "", Line: l.line, Column: l.col}
	}

	start := l.pos
	ch := l.input[start]
	col := l.col

	// Handles identifiers and keywords
	if isLetter(ch) {
		return l.readKeywordOrIdentifier()
	}

	// Handles numbers
	if unicode.IsDigit(ch) || (ch == '-' && unicode.IsDigit(l.peek())) {
		return Token{Type: TokenNumber, Lexeme: l.readNumber(), Line: l.line, Column: col}
	}

	// Handle strings
	if ch == '"' {
		l.pos++
		l.col++
		return Token{Type: TokenString, Lexeme: l.readString(), Line: l.line, Column: col}
	}

	// Handles multi-character tokens
	operator := l.readMultiCharSymbol(ch)
	if operator != nil {
		return *operator
	}

	l.pos++
	l.col++
	// Handles operators like +, -, =, etc
	if operatorType, ok := Operators[string(ch)]; ok {
		return Token{Type: TokenOperator, Lexeme: operatorType, Line: l.line, Column: col}
	}

	// Handles symbols like "(", ")", ",", ";", etc
	if symbolType, ok := Symbols[string(ch)]; ok {
		return Token{Type: symbolType, Lexeme: string(ch), Line: l.line, Column: col}
	}

	// Handles unknown tokens
	return Token{Type: TokenUnknown, Lexeme: string(ch), Line: l.line, Column: col}
}

// Handles keywords and identifiers
func (l *Lexer) readKeywordOrIdentifier() Token {
	col := l.col
	// Handles multi word keywords
	if l.matchMultiWordKeyword("trong", "khi") {
		return Token{Type: TokenKeyword, Lexeme: KeywordTrongKhi, Line: l.line, Column: col}
	}
	if l.matchMultiWordKeyword("trả", "về") {
		return Token{Type: TokenKeyword, Lexeme: KeywordTraVe, Line: l.line, Column: col}
	}
	if l.matchMultiWordKeyword("kết", "thúc") {
		return Token{Type: TokenKeyword, Lexeme: KeywordKetThuc, Line: l.line, Column: col}
	}
	if l.matchMultiWordKeyword("không", "thì") {
		return Token{Type: TokenKeyword, Lexeme: KeywordKhongThi, Line: l.line, Column: col}
	}
	if l.matchMultiWordKeyword("thủ", "tục") {
		return Token{Type: TokenKeyword, Lexeme: KeywordThuTuc, Line: l.line, Column: col}
	}

	ident := l.readIdentifier()

	// Exception for E
	if ident == "E" {
		return Token{Type: TokenOperator, Lexeme: SymbolMember, Line: l.line, Column: col}
	}

	// Handles primitive container types
	if containerType, ok := Containers[ident]; ok {
		return Token{Type: TokenContainer, Lexeme: containerType, Line: l.line, Column: col}
	}

	// Handle primitive types
	if primitiveType, ok := Primitives[ident]; ok {
		return Token{Type: TokenPrimitive, Lexeme: primitiveType, Line: l.line, Column: col}
	}

	// Handles single word keywords
	if keywordType, ok := Keywords[ident]; ok {
		return Token{Type: TokenKeyword, Lexeme: keywordType, Line: l.line, Column: col}
	}
	return Token{Type: TokenIdent, Lexeme: ident, Line: l.line, Column: col}

}

// Multi-word keyword handler
func (l *Lexer) matchMultiWordKeyword(first string, rest ...string) bool {
	if l.input[l.pos:] == nil {
		return false
	}

	startPos := l.pos
	startCol := l.col

	if l.readIdentifier() != first {
		l.pos = startPos
		l.col = startCol
		return false
	}

	for _, word := range rest {
		l.skipWhitespace()
		if l.pos >= len(l.input) || !isLetter(l.input[l.pos]) {
			l.pos = startPos
			l.col = startCol
			return false
		}
		next := l.readIdentifier()
		if next != word {
			l.pos = startPos
			l.col = startCol
			return false
		}
	}

	return true
}

// Handles multi character symbols
func (l *Lexer) readMultiCharSymbol(ch rune) *Token {
	col := l.col
	switch ch {
	case '/':
		if l.peek() == '/' {
			l.readChar() // Comsumes '/'
			l.readChar() // Consumes '/'
			return l.skipComment()
		}
	case '-':
		if l.peek() == '>' {
			l.readChar() // Consumes '-'
			l.readChar() // Consumes '>'
			return &Token{Type: TokenOperator, Lexeme: SymbolArrow, Line: l.line, Column: col}
		}
		break
	case '<':
		if l.peek() == '=' {
			l.readChar()
			l.readChar()
			return &Token{Type: TokenOperator, Lexeme: SymbolLessEqual, Line: l.line, Column: col}
		}
	case '>':
		if l.peek() == '=' {
			l.readChar()
			l.readChar()
			return &Token{Type: TokenOperator, Lexeme: SymbolGreaterEqual, Line: l.line, Column: col}
		}
	case ':':
		if l.peek() == '=' {
			l.readChar()
			l.readChar()
			return &Token{Type: TokenOperator, Lexeme: SymbolAssign, Line: l.line, Column: col}
		}
	case '!':
		if l.peek() == '=' {
			l.readChar()
			l.readChar()
			return &Token{Type: TokenOperator, Lexeme: SymbolNotEqual, Line: l.line, Column: col}
		}
	}
	return nil
}

// Identifier handler
func (l *Lexer) readIdentifier() string {
	start := l.pos
	for l.pos < len(l.input) && (isLetter(l.input[l.pos]) /*|| l.input[l.pos] == ' '*/ || unicode.IsDigit(l.input[l.pos])) {
		l.pos++
		l.col++
	}
	return string(l.input[start:l.pos])
}

func (l *Lexer) readNumber() string {
	start := l.pos
	isFloat := false
	if l.input[l.pos] == '-' {
		l.pos++
		l.col++
	}
	for l.pos < len(l.input) && (unicode.IsDigit(l.input[l.pos]) || (l.input[l.pos] == '.' && !isFloat)) {
		if l.input[l.pos] == '.' {
			isFloat = true
		}
		l.pos++
		l.col++
	}
	// This is cool
	return string(l.input[start:l.pos])
}

// String handler
func (l *Lexer) readString() string {
	start := l.pos
	// Maybe add support for \n and others
	for l.pos < len(l.input) && l.input[l.pos] != '"' {
		l.pos++
		l.col++
	}
	end := l.pos
	l.pos++
	l.col++
	return string(l.input[start:end])
}

// Skip whitespaces
func (l *Lexer) skipWhitespace() *Token {
	for l.pos < len(l.input) && unicode.IsSpace(l.input[l.pos]) {
		if l.input[l.pos] == '\n' {
			col, line := l.col, l.line
			l.line++
			l.col = 1
			l.pos++
			return &Token{Type: TokenNewLine, Lexeme: "\\n", Line: line, Column: col}
		}
		l.pos++
		l.col++
	}
	return nil
}

func (l *Lexer) skipComment() *Token {
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.pos++
		l.col++
	}
	col, line := l.col, l.line
	l.line++
	l.pos++
	l.col = 1
	return &Token{Type: TokenNewLine, Lexeme: "\\n", Line: line, Column: col}
}

// Letter handler
func isLetter(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

// Safely checks the next character
func (l *Lexer) peek() rune {
	if l.pos+1 >= len(l.input) {
		return rune(-1)
	}
	return l.input[l.pos+1]
}

// Reads and consumes next character
func (l *Lexer) readChar() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	ch := l.input[l.pos]
	l.pos++
	// Probably doesn't happen
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}
