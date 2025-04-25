package main

type Parser struct {
	tokens  []Token
	pos     int
	current Token
}

func NewParser(tokens []Token) *Parser {
	p := &Parser{tokens: tokens, pos: -1}
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.pos++
	if p.pos < len(p.tokens) {
		p.current = p.tokens[p.pos]
	} else {
		p.current = Token{Type: TokenEOF}
	}
}

func (p *Parser) peekToken() Token {
	if p.pos+1 < len(p.tokens) {
		return p.tokens[p.pos+1]
	}
	return Token{Type: TokenEOF}
}

/* Probably not gonna use
func (p *Parser) expect(t TokenType, args ...string) (Token, error) {
	if p.current.Type == t {
		tok := p.current
		p.nextToken()
		return tok, nil
	}
	return Token{}, NewLangError(WrongToken, args)
}
*/

func (p *Parser) ParseProgram() (*Program, error) {
	prog := &Program{}

	// Handles end of file
	for p.current.Type != TokenEOF {
		// Skip new lines
		for p.current.Type == TokenNewLine {
			p.nextToken()
		}
		fn, err := p.parseFunction()
		if err != nil {
			return nil, err
		}
		prog.Functions = append(prog.Functions, fn)
		// Skip new lines
		for p.current.Type == TokenNewLine {
			p.nextToken()
		}
	}

	return prog, nil
}

func (p *Parser) parseFunction() (*Function, error) {
	// Expect 'hàm' keyword
	if p.current.Type != TokenKeyword || p.current.Lexeme != KeywordHam {
		return nil, NewLangError(WrongToken, "hàm", p.current.Lexeme).At(p.current.Line, p.current.Column)
	}
	p.nextToken() // Consumes 'hàm'

	// Expect function name (identifier)
	if p.current.Type != TokenIdent {
		return nil, NewLangError(ExpectToken, "tên hàm").At(p.current.Line, p.current.Column)
	}
	fnName := p.current.Lexeme
	line, col := p.current.Line, p.current.Column
	p.nextToken()

	// Handles parameters
	// Expect '('
	if p.current.Type != TokenLParen {
		return nil, NewLangError(WrongToken, "(", p.current.Lexeme).At(p.current.Line, p.current.Column)
	}
	p.nextToken() // Consumes '('
	params := []*Variable{}
	if p.current.Type != TokenRParen {
		for {
			if p.current.Type != TokenIdent {
				return nil, NewLangError(WrongToken, "tên tham số", p.current.Lexeme).At(p.current.Line, p.current.Column)
			}
			line := p.current.Line
			col := p.current.Column
			paramName := p.current.Lexeme
			p.nextToken()
			if p.current.Type != TokenOperator || p.current.Lexeme != SymbolMember {
				return nil, NewLangError(WrongToken, SymbolMember, p.current.Lexeme).At(p.current.Line, p.current.Column)
			}
			p.nextToken() // Consumes the 'E'

			if p.current.Type != TokenPrimitive && p.current.Type != TokenIdent {
				return nil, NewLangError(ExpectToken, "kiểu dữ liệu").At(p.current.Line, p.current.Column)
			}
			paramType := p.current.Lexeme
			p.nextToken()

			params = append(params, &Variable{Name: paramName, Type: paramType, Line: line, Column: col})

			if p.current.Type == TokenComma {
				p.nextToken()
				continue
			}

			if p.current.Type != TokenRParen {
				return nil, NewLangError(WrongToken, ")", p.current.Lexeme).At(p.current.Line, p.current.Column)
			}
			break
		}
	}
	p.nextToken() // Consumes ')'

	// TODO: Parse return type (optional)
	// ...
	returnType := "rỗng"
	if p.current.Type == TokenOperator && p.current.Lexeme == SymbolArrow {
		p.nextToken() // Consumes '->'
		if p.current.Type != TokenIdent && p.current.Type != TokenPrimitive {
			return nil, NewLangError(ExpectToken, "kiểu trả về").At(p.current.Line, p.current.Column)
		}
		returnType = p.current.Lexeme
		p.nextToken() // Consumes the type
	}

	// Forces you to create a new line
	if p.current.Type != TokenNewLine {
		return nil, NewLangError(ExpectToken, "xuống dòng").At(p.current.Line, p.current.Column)
	}
	p.nextToken()

	// Parse function body statements until 'kết thúc'
	body := []Statement{}
	for !(p.current.Type == TokenKeyword && p.current.Lexeme == KeywordKetThuc) && p.current.Type != TokenEOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		body = append(body, stmt)
		if p.current.Type != TokenNewLine && p.current.Type != TokenSemiColon {
			return nil, NewLangError(ExpectToken, "xuống dòng hoặc ';'").At(p.current.Line, p.current.Column)
		}
		p.nextToken()
	}

	// Expect 'kết thúc'
	if p.current.Type != TokenKeyword || p.current.Lexeme != KeywordKetThuc {
		return nil, NewLangError(WrongToken, "kết thúc", p.current.Lexeme).At(p.current.Line, p.current.Column)
	}
	p.nextToken()

	return &Function{
		Name:       fnName,
		Parameters: params,
		ReturnType: returnType,
		Body:       body,
		Line:       line,
		Column:     col,
	}, nil
}

func (p *Parser) parseStatement() (Statement, error) {
	for p.current.Type == TokenNewLine {
		p.nextToken()
	}
	switch p.current.Type {
	case TokenKeyword:
		switch p.current.Lexeme {
		case KeywordBien: // variable declartion
			return p.parseVarDecl()
		case KeywordTraVe: // return statement
			return p.parseReturnStmt()
		default:
			return nil, NewLangError(UnexpectedToken, p.current.Lexeme).At(p.current.Line, p.current.Column)
		}
	default:
		return nil, NewLangError(UnexpectedToken, p.current.Lexeme).At(p.current.Line, p.current.Column)
	}
}

// TODO: Add multiple variable declaration in one line (its one per line for simplicity for now...)
func (p *Parser) parseVarDecl() (Statement, error) {
	// Consume 'biến'
	p.nextToken()

	// Expect identifier
	if p.current.Type != TokenIdent {
		return nil, NewLangError(ExpectToken, "tên biến").At(p.current.Line, p.current.Column)
	}
	varName := p.current.Lexeme
	line, col := p.current.Line, p.current.Column
	p.nextToken()

	// Type declaration
	// TODO: infer type annotation later
	if p.current.Type != TokenOperator || p.current.Lexeme != SymbolMember {
		return nil, NewLangError(WrongToken, SymbolMember, p.current.Lexeme).At(p.current.Line, p.current.Column)
	}
	p.nextToken() // Consumes "E"
	if p.current.Type != TokenPrimitive && p.current.Type != TokenIdent {
		return nil, NewLangError(ExpectToken, "kiểu dữ liệu").At(p.current.Line, p.current.Column)
	}
	varType := p.current.Lexeme
	p.nextToken()

	// optional: assignment
	// For example: ":= expression"
	if p.current.Type == TokenOperator && p.current.Lexeme == SymbolAssign {
		p.nextToken()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		// TODO: Handle type mismatch
		return &VarDecl{
			Var:    &Variable{Name: varName, Type: varType, Line: line, Column: col},
			Value:  expr,
			Line:   line,
			Column: col,
		}, nil
	}

	// No initializer found
	return &VarDecl{
		Var:    &Variable{Name: varName, Type: varType, Line: line, Column: col},
		Value:  nil,
		Line:   line,
		Column: col,
	}, nil
}

func (p *Parser) parseReturnStmt() (Statement, error) {
	// consume 'trả về'
	p.nextToken()
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	line, column := expr.Pos()
	return &ReturnStmt{Value: expr, Line: line, Column: column}, nil
}

// Basic implementaion
// TODO: Make more advanced expression parser
func (p *Parser) parseExpression() (Expression, error) {
	switch p.current.Type {
	case TokenIdent:
		id := &Identifier{Name: p.current.Lexeme, Line: p.current.Line, Column: p.current.Column}
		p.nextToken()
		return id, nil
	case TokenNumber:
		// TODO: get number type for real (for now its R64)
		num := &NumberLiteral{Value: p.current.Lexeme, Type: PrimitiveR64, Line: p.current.Line, Column: p.current.Column}
		p.nextToken()
		return num, nil
	default:
		return nil, NewLangError(UnexpectedToken, p.current.Lexeme).At(p.current.Line, p.current.Column)
	}
}
