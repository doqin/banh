package main

import "strings"

var precedences = map[string]int{
	KeywordHoac:        3,
	KeywordVa:          6,
	SymbolEqual:        12,
	SymbolNotEqual:     12,
	SymbolLessEqual:    25,
	SymbolLess:         25,
	SymbolGreaterEqual: 25,
	SymbolGreater:      25,
	SymbolPlus:         50,
	SymbolMinus:        50,
	SymbolAsterisk:     100,
	SymbolSlash:        100,
}

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
		switch p.current.Lexeme {
		case KeywordHam:
			fn, err := p.parseFunction()
			if err != nil {
				return nil, err
			}
			prog.Functions = append(prog.Functions, fn)
		case KeywordThuTuc:
			fn, err := p.parseProcedure()
			if err != nil {
				return nil, err
			}
			prog.Functions = append(prog.Functions, fn)
		default:
			return nil, NewLangError(UnexpectedToken, p.current.Lexeme).At(p.current.Line, p.current.Column)
		}

		// Skip new lines
		for p.current.Type == TokenNewLine {
			p.nextToken()
		}
	}

	return prog, nil
}

func (p *Parser) parseProcedure() (*Function, error) {
	// Expect 'thủ tục' keyword
	if p.current.Type != TokenKeyword || p.current.Lexeme != KeywordThuTuc {
		return nil, NewLangError(WrongToken, KeywordThuTuc, p.current.Lexeme).At(p.current.Line, p.current.Column)
	}
	p.nextToken() // Consumes 'thủ tục'

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
			line := p.current.Line
			col := p.current.Column
			paramName, paramType, err := p.parseVarIdent()
			if err != nil {
				return nil, err
			}
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
		// FIXME: Implement return statement without returning a value
		st, ok := stmt.(*ReturnStmt)
		if ok && st.Value != nil {
			return nil, NewLangError(ReturnTypeMismatch, getExprType(st.Value).String(), PrimitiveVoid)
		}
		body = append(body, stmt)
		if p.current.Type != TokenNewLine && p.current.Type != TokenSemiColon {
			return nil, NewLangError(ExpectToken, "xuống dòng hoặc ';'").At(p.current.Line, p.current.Column)
		}
		p.nextToken()
		for p.current.Type == TokenNewLine {
			p.nextToken()
		}
	}

	// Expect 'kết thúc'
	if p.current.Type != TokenKeyword || p.current.Lexeme != KeywordKetThuc {
		return nil, NewLangError(WrongToken, "kết thúc", p.current.Lexeme).At(p.current.Line, p.current.Column)
	}
	p.nextToken()

	return &Function{
		Name:       fnName,
		Parameters: params,
		ReturnType: &PrimitiveType{Name: PrimitiveVoid},
		Body:       body,
		Line:       line,
		Column:     col,
	}, nil
}

func (p *Parser) parseFunction() (*Function, error) {
	// Expect 'hàm' keyword
	if p.current.Type != TokenKeyword || p.current.Lexeme != KeywordHam {
		return nil, NewLangError(WrongToken, KeywordHam, p.current.Lexeme).At(p.current.Line, p.current.Column)
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
			line := p.current.Line
			col := p.current.Column
			paramName, paramType, err := p.parseVarIdent()
			if err != nil {
				return nil, err
			}
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

	if p.current.Type != TokenOperator || p.current.Lexeme != SymbolArrow {
		return nil, NewLangError(WrongToken, "->", p.current.Lexeme).At(p.current.Line, p.current.Column)
	}
	p.nextToken() // Consumes '->'
	if p.current.Type != TokenIdent && p.current.Type != TokenPrimitive {
		return nil, NewLangError(ExpectToken, "kiểu trả về").At(p.current.Line, p.current.Column)
	}
	returnType, err := p.parseType()
	if err != nil {
		return nil, err
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
		for p.current.Type == TokenNewLine {
			p.nextToken()
		}
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
		case KeywordNeu:
			return p.parseIfStmt()
		case KeywordBien: // variable declartion
			return p.parseVarDecl()
		case KeywordTraVe: // return statement
			return p.parseReturnStmt()
		default:
			return nil, NewLangError(UnexpectedToken, p.current.Lexeme).At(p.current.Line, p.current.Column)
		}
	case TokenIdent:
		return p.parseRegExpr()
	default:
		return nil, NewLangError(UnexpectedToken, p.current.Lexeme).At(p.current.Line, p.current.Column)
	}
}

func (p *Parser) parseRegExpr() (Statement, error) {
	line, col := p.current.Line, p.current.Column
	expr, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}
	return &RegExpr{
		Expr:   expr,
		Line:   line,
		Column: col,
	}, nil
}

// TODO: Add multiple variable declaration in one line (its one per line for simplicity for now...)
func (p *Parser) parseVarDecl() (Statement, error) {
	line, col := p.current.Line, p.current.Column
	// Consume 'biến'
	p.nextToken()

	varName, varType, err := p.parseVarIdent()
	if err != nil {
		return nil, err
	}
	// Optional: assignment
	// For example: ":= expression"
	if p.current.Type == TokenOperator && p.current.Lexeme == SymbolAssign {
		p.nextToken()
		switch varType.(type) {
		case *PrimitiveType, *ContainerType:
			expr, err := p.parseExpression(0)
			if err != nil {
				return nil, err
			}
			return &VarDecl{
				Var:    &Variable{Name: varName, Type: varType, Line: line, Column: col},
				Value:  expr,
				Line:   line,
				Column: col,
			}, nil
		case *StructType:
			// TODO: Implement it
			panic("Chưa cài đặt kiểu dữ liệu có cấu trúc")
		default:
			panic("Không nhận dạng được kiểu dữ liệu")
		}
	}

	// No initializer found
	return &VarDecl{
		Var:    &Variable{Name: varName, Type: varType, Line: line, Column: col},
		Value:  &UninitializedExpr{Line: p.current.Line, Column: p.current.Column},
		Line:   line,
		Column: col,
	}, nil
}

func (p *Parser) parseArray() (*ArrayLiteral, error) {
	line, col := p.current.Line, p.current.Column
	if p.current.Lexeme != "{" {
		return nil, NewLangError(ExpectToken, "{").At(p.current.Line, p.current.Column)
	}
	p.nextToken() // Consumes "{"
	elements := []Expression{}
	for p.current.Lexeme != "}" {
		element, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}
		elements = append(elements, element)
		if p.current.Lexeme != "}" && p.current.Lexeme != "," {
			return nil, NewLangError(ExpectToken, "}").At(p.current.Line, p.current.Column)
		}
		if p.current.Lexeme == "}" {
			break
		}
		p.nextToken() // Consumes "}" or ","
	}
	p.nextToken()
	return &ArrayLiteral{Elements: elements, Type: &UnknownType{Name: "Unknown"}, Line: line, Column: col}, nil
}

func (p *Parser) parseVarIdent() (string, Type, error) {
	// Checks for identifier
	if p.current.Type != TokenIdent {
		// TODO: Change accordingly
		return "", nil, NewLangError(WrongToken, "tên biến", p.current.Lexeme).At(p.current.Line, p.current.Column)
	}
	varName := p.current.Lexeme
	p.nextToken()

	if p.current.Type != TokenOperator || p.current.Lexeme != SymbolMember {
		return "", nil, NewLangError(WrongToken, SymbolMember, p.current.Lexeme).At(p.current.Line, p.current.Column)
	}
	p.nextToken() // Consumes the 'E'

	varType, err := p.parseType()
	if err != nil {
		return "", nil, err
	}
	return varName, varType, nil
}

func (p *Parser) parseType() (Type, error) {
	switch p.current.Type {
	case TokenPrimitive:
		varType := &PrimitiveType{Name: p.current.Lexeme}
		p.nextToken()
		return varType, nil
	case TokenIdent:
		varType := &StructType{Name: p.current.Lexeme}
		p.nextToken()
		return varType, nil
	case TokenContainer:
		containerKind := p.current.Lexeme
		p.nextToken()

		// Get dimension from container type
		var dimension int
		switch containerKind {
		case ContainerArray, ContainerHashMap:
			dimension = 1
		case ContainerMatrix:
			dimension = 2
		default:
			return nil, NewLangError(ExpectToken, "kiểu dữ liệu").At(p.current.Line, p.current.Column)
		}
		var bounds []Expression

		// Get the bounds
		for range dimension {
			if p.current.Type != TokenLBrack {
				return nil, NewLangError(WrongToken, "[", p.current.Lexeme).At(p.current.Line, p.current.Column)
			}
			p.nextToken()
			leftBound, err := p.parseExpression(0) // Get Left Bound
			if err != nil {
				return nil, err
			}
			// Handle left bound type
			typeInterface := getExprType(leftBound)
			typ, ok := typeInterface.(*PrimitiveType)
			if (ok && typ.Name == PrimitiveR32 || typ.Name == PrimitiveR64) || !ok {
				return nil, NewLangError(ExpectToken, "số tự nhiên hoặc số nguyên").At(p.current.Line, p.current.Column)
			}
			if p.current.Lexeme != SymbolDotDot {
				return nil, NewLangError(WrongToken, SymbolDotDot, p.current.Lexeme).At(p.current.Line, p.current.Column)
			}
			p.nextToken()
			rightBound, err := p.parseExpression(0) // Get Right Bound
			if err != nil {
				return nil, err
			}
			// Handle right bound type
			typeInterface = getExprType(rightBound)
			typ, ok = typeInterface.(*PrimitiveType)
			if (ok && typ.Name == PrimitiveR32 || typ.Name == PrimitiveR64) || !ok {
				return nil, NewLangError(ExpectToken, "số tự nhiên hoặc số nguyên").At(p.current.Line, p.current.Column)
			}
			// Close the range expression
			if p.current.Type != TokenRBrack {
				return nil, NewLangError(WrongToken, "]", p.current.Lexeme).At(p.current.Line, p.current.Column)
			}
			p.nextToken() // Consumes the ']'
			bounds = append(bounds, leftBound)
			bounds = append(bounds, rightBound)
		}
		if p.current.Type != TokenOperator || p.current.Lexeme != SymbolMember {
			return nil, NewLangError(WrongToken, SymbolMember, p.current.Lexeme).At(p.current.Line, p.current.Column)
		}
		p.nextToken() // Consumes the 'E'
		elementType, err := p.parseType()
		if err != nil {
			return nil, err
		}
		varType := &ContainerType{Kind: containerKind, ElementType: elementType, Dimensions: dimension, Bounds: bounds}
		return varType, nil
	default:
		return nil, NewLangError(ExpectToken, "kiểu dữ liệu").At(p.current.Line, p.current.Column)
	}
}

func (p *Parser) parseReturnStmt() (Statement, error) {
	line, column := p.current.Line, p.current.Column
	// consume 'trả về'
	p.nextToken()
	if p.current.Type == TokenNewLine || p.current.Lexeme == PrimitiveVoid || p.current.Type == TokenSemiColon {
		p.nextToken()
		return &ReturnStmt{Value: nil, Line: line, Column: column}, nil
	}
	expr, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}
	return &ReturnStmt{Value: expr, Line: line, Column: column}, nil
}

func (p *Parser) parseIfStmt() (Statement, error) {
	line, column := p.current.Line, p.current.Column
	// Consumes 'nếu'
	p.nextToken()
	condi, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}
	if p.current.Type != TokenKeyword || p.current.Lexeme != KeywordThi {
		return nil, NewLangError(WrongToken, TokenKeyword, p.current.Lexeme).At(p.current.Line, p.current.Column)
	}
	p.nextToken() // Consumes the 'thì'
	thenBlock := []Statement{}
	for p.current.Type != TokenKeyword && (p.current.Lexeme != KeywordKetThuc && p.current.Lexeme != KeywordKhongThi) && p.current.Type != TokenEOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		thenBlock = append(thenBlock, stmt)
		if p.current.Type != TokenNewLine && p.current.Type != TokenSemiColon {
			return nil, NewLangError(ExpectToken, "xuống dòng hoặc ';'").At(p.current.Line, p.current.Column)
		}
		p.nextToken()
		for p.current.Type == TokenNewLine {
			p.nextToken()
		}
	}
	switch p.current.Lexeme {
	case KeywordKetThuc:
		p.nextToken() // Consumes 'kết thúc'
		return &IfStmt{
			Condition: condi,
			ThenBlock: thenBlock,
			ElseBlock: nil,
			Line:      line,
			Column:    column,
		}, nil
	case KeywordKhongThi:
		p.nextToken() // Consumes 'không thì'

		switch p.current.Lexeme {
		case KeywordNeu:
			elseBlock, err := p.parseIfStmt()
			if err != nil {
				return nil, err
			}
			return &IfStmt{
				Condition: condi,
				ThenBlock: thenBlock,
				ElseBlock: []Statement{elseBlock},
				Line:      line,
				Column:    column,
			}, nil
		default:
			if p.current.Type != TokenNewLine {
				return nil, NewLangError(ExpectToken, "xuống dòng").At(p.current.Line, p.current.Column)
			}
			p.nextToken() // Consumes '\n'
			elseBlock := []Statement{}
			for p.current.Lexeme != KeywordKetThuc && p.current.Type != TokenEOF {
				stmt, err := p.parseStatement()
				if err != nil {
					return nil, err
				}
				elseBlock = append(elseBlock, stmt)
				if p.current.Type != TokenNewLine && p.current.Type != TokenSemiColon {
					return nil, NewLangError(ExpectToken, "xuống dòng hoặc ';'").At(p.current.Line, p.current.Column)
				}
				p.nextToken()
				for p.current.Type == TokenNewLine {
					p.nextToken()
				}
			}
			if p.current.Lexeme != KeywordKetThuc {
				return nil, NewLangError(WrongToken, KeywordKetThuc, p.current.Lexeme).At(p.current.Line, p.current.Column)
			}
			p.nextToken() // Consumes 'kết thúc'
			return &IfStmt{
				Condition: condi,
				ThenBlock: thenBlock,
				ElseBlock: elseBlock,
				Line:      line,
				Column:    column,
			}, nil
		}
	default:
		return nil, NewLangError(WrongToken, KeywordKetThuc, p.current.Lexeme).At(p.current.Line, p.current.Column)
	}
}

func (p *Parser) parseCallExpr() (Expression, error) {
	line, column := p.current.Line, p.current.Column
	fnName := p.current.Lexeme
	p.nextToken() // Consumes name
	if p.current.Type != TokenLParen {
		return nil, NewLangError(WrongToken, "(", p.current.Lexeme).At(p.current.Line, p.current.Column)
	}
	p.nextToken() // Consumes '('
	var arguments []Expression
	if p.current.Type != TokenRParen {
		for {
			expr, err := p.parseExpression(0)
			if err != nil {
				return nil, err
			}
			arguments = append(arguments, expr)
			if p.current.Type == TokenRParen {
				break
			}
			if p.current.Type != TokenComma {
				return nil, NewLangError(WrongToken, "(", p.current.Lexeme).At(p.current.Line, p.current.Column)
			}
			p.nextToken()
		}
	}
	p.nextToken() // Consumes ')'
	return &CallExpr{Name: fnName, Arguments: arguments, ReturnType: &UnknownType{Name: "Unknown"}, Line: line, Column: column}, nil
}

func (p *Parser) parseExplicitCast() (Expression, error) {
	line, column := p.current.Line, p.current.Column
	castType := p.current.Lexeme // Primitive type
	p.nextToken()                // Consumes the type
	if p.current.Type != TokenLParen {
		return nil, NewLangError(WrongToken, "(", p.current.Lexeme).At(p.current.Line, p.current.Column)
	}
	p.nextToken() // Consumes '('
	argument, err := p.parseExpression(0)
	if err != nil {
		return nil, err
	}
	if p.current.Type != TokenRParen {
		return nil, NewLangError(WrongToken, ")", p.current.Lexeme).At(p.current.Line, p.current.Column)
	}
	p.nextToken() // Consumes ')'
	return &ExplicitCast{Type: PrimitiveType{Name: castType}, Argument: argument, Line: line, Column: column}, nil
}

// TODO: Make more advanced expression parser
func (p *Parser) parseExpression(minPrec int) (Expression, error) {
	// Parse the left-hand side (primary expression)
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	// Parse binary operators according to precedence
	for {
		prec := p.currentPrecedence()
		if prec < minPrec {
			break
		}

		op := p.current.Lexeme
		line, col := p.current.Line, p.current.Column
		p.nextToken() // consume operator

		// Parse right-hand side expression with higher precedence for right-associativity
		right, err := p.parseExpression(prec + 1)
		if err != nil {
			return nil, err
		}

		left = &BinaryExpr{
			Left:       left,
			Operator:   op,
			Right:      right,
			ReturnType: PrimitiveType{Name: "Unknown"}, // Placeholder
			Line:       line,
			Column:     col,
		}
	}

	return left, nil
}

func (p *Parser) parsePrimary() (Expression, error) {
	switch p.current.Type {
	case TokenPrimitive:
		casted, err := p.parseExplicitCast()
		if err != nil {
			return nil, err
		}
		return casted, nil
	case TokenIdent:
		if p.peekToken().Type == TokenLParen {
			call, err := p.parseCallExpr()
			if err != nil {
				return nil, err
			}
			return call, nil
		} else {
			id := &Identifier{Name: p.current.Lexeme, Type: &UnknownType{Name: "Unknown"}, Line: p.current.Line, Column: p.current.Column}
			p.nextToken()
			return id, nil
		}
	case TokenNumber:
		var num *NumberLiteral
		if strings.Contains(p.current.Lexeme, ".") {
			num = &NumberLiteral{Value: p.current.Lexeme, Type: PrimitiveType{Name: PrimitiveR64}, Line: p.current.Line, Column: p.current.Column}
		} else {
			num = &NumberLiteral{Value: p.current.Lexeme, Type: PrimitiveType{Name: PrimitiveZ64}, Line: p.current.Line, Column: p.current.Column}
		}
		p.nextToken()
		return num, nil
	case TokenLParen:
		p.nextToken() // Consumes '('
		expr, err := p.parseExpression(0)
		if err != nil {
			return nil, err
		}
		if p.current.Type != TokenRParen {
			return nil, NewLangError(WrongToken, ")", p.current.Lexeme).At(p.current.Line, p.current.Column)
		}
		p.nextToken() // Consumes ')'
		return expr, nil
	case TokenLBrace:
		expr, err := p.parseArray()
		if err != nil {
			return nil, err
		}
		return expr, nil
	default:
		return nil, NewLangError(UnexpectedToken, p.current.Lexeme).At(p.current.Line, p.current.Column)
	}
}

// Helper function
func (p *Parser) currentPrecedence() int {
	if p.current.Type == TokenOperator || p.current.Type == TokenKeyword {
		if prec, ok := precedences[p.current.Lexeme]; ok {
			return prec
		}
	}
	return -1
}

/*
// Helper function
func isTypeNumber_Expr(expr Expression) bool {
	switch expr := expr.(type) {
	case *NumberLiteral:
		switch expr.Type.Name {
		case PrimitiveN32, PrimitiveN64, PrimitiveZ32, PrimitiveZ64, PrimitiveR32, PrimitiveR64:
			return true
		default:
			return false
		}
	case *BinaryExpr:
		switch expr.ReturnType.Name {
		case PrimitiveN32, PrimitiveN64, PrimitiveZ32, PrimitiveZ64, PrimitiveR32, PrimitiveR64:
			return true
		default:
			return false
		}
	case *Identifier:
		switch expr := expr.Type.(type) {
		case *PrimitiveType:
			switch expr.Name {
			case PrimitiveN32, PrimitiveN64, PrimitiveZ32, PrimitiveZ64, PrimitiveR32, PrimitiveR64:
				return true
			default:
				return false
			}
		default:
			return false
		}
	case *CallExpr:
		switch expr := expr.ReturnType.(type) {
		case *PrimitiveType:
			switch expr.Name {
			case PrimitiveN32, PrimitiveN64, PrimitiveZ32, PrimitiveZ64, PrimitiveR32, PrimitiveR64:
				return true
			default:
				return false
			}
		default:
			return false
		}
	case *ExplicitCast:
		switch expr.Type.Name {
		case PrimitiveN32, PrimitiveN64, PrimitiveZ32, PrimitiveZ64, PrimitiveR32, PrimitiveR64:
			return true
		default:
			return false
		}
	default:
		return false
	}
}
*/

func isTypeNumber_Type(typ Type) bool {
	switch typ := typ.(type) {
	case *PrimitiveType:
		switch typ.Name {
		case PrimitiveN32, PrimitiveN64, PrimitiveZ32, PrimitiveZ64, PrimitiveR32, PrimitiveR64:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

// Helper function
func getExprType(expr Expression) Type {
	switch e := expr.(type) {
	case *Identifier:
		return e.Type
	case *NumberLiteral:
		return &e.Type
	case *BinaryExpr:
		return &e.ReturnType
	case *CallExpr:
		return e.ReturnType
	case *ExplicitCast:
		return &e.Type
	default:
		return &UnknownType{Name: "Unknown"}
	}
}
