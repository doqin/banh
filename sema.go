package main

import (
	"fmt"
	"strings"
)

// Scope represents a symbol table with optional parent scoping
type Scope struct {
	Parent    *Scope
	Symbols   map[string]*Variable
	Functions map[string]*Function
}

// NewScope creates a new scope, optionally with a parent
func NewScope(parent *Scope) *Scope {
	return &Scope{
		Parent:    parent,
		Symbols:   make(map[string]*Variable),
		Functions: make(map[string]*Function),
	}
}

// Declare adds a new variable to the current scope
// It panics if the name already exists in the current scope.
func (s *Scope) Declare(name string, v any) error {
	switch typ := v.(type) {
	case *Variable:
		if _, exists := s.Symbols[name]; exists {
			return NewLangError(RedeclarationVar, name).At(typ.Line, typ.Column)
		}
		s.Symbols[name] = typ
		return nil
	case *Function:
		if _, exists := s.Functions[name]; exists {
			return NewLangError(RedeclarationFunction, name).At(typ.Line, typ.Column)
		}
		s.Functions[name] = typ
		return nil
	default:
		return fmt.Errorf("Câu lệnh khai báo không xác định")
	}
}

// Resolve looks for a variable in the current scope chain.
func (s *Scope) Resolve(name string) (any, bool) {
	// Check for function first
	if f, ok := s.Functions[name]; ok {
		return f, true
	}
	if v, ok := s.Symbols[name]; ok {
		return v, true
	}
	// Otherwise check for parent's scope
	if s.Parent != nil {
		return s.Parent.Resolve(name)
	}
	return nil, false // Not found
}

type TypeChecker struct {
	GlobalScope  *Scope
	CurrentScope *Scope
}

// Entry point
func (tc *TypeChecker) AnalyzeProgram(p *Program) error {
	tc.GlobalScope = NewScope(nil)
	err := tc.InitializeBuiltins()
	if err != nil {
		return err
	}

	// First, declare all functions (for forward reference)
	for _, fn := range p.Functions {
		err := tc.GlobalScope.Declare(fn.Name, &Function{
			Name:       fn.Name,
			Parameters: fn.Parameters,
			ReturnType: fn.ReturnType,
			Line:       fn.Line,
			Column:     fn.Column,
		})
		if err != nil {
			return err
		}
	}

	// Then check function bodies
	for _, fn := range p.Functions {
		err := tc.AnalyzeFunction(fn)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tc *TypeChecker) InitializeBuiltins() error {
	// Define the print function signature: in(tuỳ) -> rỗng
	printFn := &Function{
		Name:       "in",
		Parameters: []*Variable{{Name: "giá_trị", Type: PrimitiveAny}},
		ReturnType: PrimitiveVoid,
		Line:       0,
		Column:     0,
	}
	err := tc.GlobalScope.Declare("in", printFn)
	if err != nil {
		return err
	}

	// TODO: Add more later
	return nil
}

func (tc *TypeChecker) AnalyzeFunction(fn *Function) error {
	tc.CurrentScope = NewScope(tc.GlobalScope)

	// Declare parameters
	for _, param := range fn.Parameters {
		err := tc.CurrentScope.Declare(param.Name, param)
		if err != nil {
			return err
		}
	}

	// Analyze body statements
	for _, stmt := range fn.Body {
		err := tc.AnalyzeStatement(stmt, fn.ReturnType)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tc *TypeChecker) AnalyzeStatement(stmt Statement, expectedReturnType string) error {
	switch s := stmt.(type) {
	case *VarDecl:
		err := tc.AnalyzeExpression(s.Value)
		if err != nil {
			return err
		}
		valType := tc.getExprType(s.Value)

		if valType != s.Var.Type && isLiteral(s.Value) {
			if !canLiteralCast(valType, s.Var.Type) {
				line, col := s.Pos()
				return NewLangError(TypeMismatch, valType, s.Var.Type).At(line, col)
			}
			valType = s.Var.Type
		}

		if valType != s.Var.Type && !canImplicitCast(valType, s.Var.Type) {
			line, col := s.Pos()
			return NewLangError(TypeMismatch, valType, s.Var.Type).At(line, col)
		}
		castExpr(s.Value, s.Var.Type)
		err = tc.CurrentScope.Declare(s.Var.Name, s.Var)
		if err != nil {
			return err
		}
		return nil
	case *ReturnStmt:
		err := tc.AnalyzeExpression(s.Value)
		if err != nil {
			return err
		}
		valType := tc.getExprType(s.Value)

		if valType != expectedReturnType && isLiteral(s.Value) {
			if !canLiteralCast(valType, expectedReturnType) {
				line, col := s.Pos()
				return NewLangError(ReturnTypeMismatch, valType, expectedReturnType).At(line, col)
			}
			valType = expectedReturnType
		}

		if valType != expectedReturnType && !canImplicitCast(valType, expectedReturnType) {
			line, col := s.Pos()
			return NewLangError(ReturnTypeMismatch, valType, expectedReturnType).At(line, col)
		}
		castExpr(s.Value, expectedReturnType)
		return nil
	case *IfStmt:
		err := tc.AnalyzeExpression(s.Condition)
		if err != nil {
			return err
		}
		for _, stmt := range s.ThenBlock {
			err := tc.AnalyzeStatement(stmt, expectedReturnType)
			if err != nil {
				return err
			}
		}
		if s.ElseBlock != nil {
			for _, stmt = range s.ElseBlock {
				err := tc.AnalyzeStatement(stmt, expectedReturnType)
				if err != nil {
					return err
				}
			}
		}
		return nil
	case *RegExpr:
		err := tc.AnalyzeExpression(s.Expr)
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("Câu lệnh không xác định")
	}
}

func (tc *TypeChecker) AnalyzeExpression(expr Expression) error {
	switch e := expr.(type) {
	case *Identifier:
		err := tc.AnalyzeIdentifier(e)
		if err != nil {
			return err
		}
		return nil
	case *NumberLiteral:
		// Is already R64
		return nil
	case *BinaryExpr:
		err := tc.AnalyzeBinaryExpr(e)
		if err != nil {
			return err
		}
		return nil
	case *CallExpr:
		err := tc.AnalyzeCallExpr(e)
		if err != nil {
			return err
		}
		return nil
	case *ExplicitCast:
		err := tc.AnalyzeExplicitCast(e)
		if err != nil {
			return err
		}
		return nil
	default:
		line, col := e.Pos()
		return NewLangError(UnknownExpression).At(line, col)
	}
}

func (tc *TypeChecker) AnalyzeExplicitCast(e *ExplicitCast) error {
	castType := e.Type
	err := tc.AnalyzeExpression(e.Argument)
	if err != nil {
		return err
	}
	exprType := tc.getExprType(e.Argument)
	if !canExplicitCast(exprType, castType) {
		return fmt.Errorf("[Dòng %d, Cột %d] Không thể chuyển kiểu '%s' sang kiểu '%s'", e.Line, e.Column, exprType, castType)
	}
	// castExpr(e.Argument, castType)
	return nil
}

func (tc *TypeChecker) AnalyzeIdentifier(i *Identifier) error {
	v, found := tc.CurrentScope.Resolve(i.Name)
	if !found {
		line, col := i.Pos()
		return NewLangError(UndeclaredIdentifier, i.Name).At(line, col)
	}

	switch typ := v.(type) {
	case *Variable:
		i.Type = typ.Type
		return nil
	case *Function:
		//FIXME: Handle function name clashing with variable name
		// Rather important
		line, col := i.Pos()
		return NewLangError(InvalidIdentifierUsage, i.Name).At(line, col)
	default:
		line, col := i.Pos()
		return NewLangError(UnknownIdentifierType).At(line, col)
	}
}

func (tc *TypeChecker) AnalyzeBinaryExpr(b *BinaryExpr) error {
	err := tc.AnalyzeExpression(b.Left)
	if err != nil {
		return err
	}
	err = tc.AnalyzeExpression(b.Right)
	if err != nil {
		return err
	}
	leftType := tc.getExprType(b.Left)
	rightType := tc.getExprType(b.Right)

	if b.Operator == KeywordVa || b.Operator == KeywordHoac {
		if leftType != PrimitiveB1 {
			return NewLangError(TypeMismatch, leftType, PrimitiveB1).At(b.Line, b.Column)
		}
		if rightType != PrimitiveB1 {
			return NewLangError(TypeMismatch, rightType, PrimitiveB1).At(b.Line, b.Column)
		}
		b.ReturnType = PrimitiveB1
		return nil
	}

	if leftType == rightType {
		b.ReturnType = leftType
		return nil
	}

	if isLiteral(b.Left) && !isLiteral(b.Right) {
		if !canLiteralCast(leftType, rightType) {
			return NewLangError(TypeMismatch, leftType, rightType).At(b.Line, b.Column)
		}
		castExpr(b.Left, rightType)
		leftType = rightType
	} else if !isLiteral(b.Left) && isLiteral(b.Right) {
		if !canLiteralCast(rightType, leftType) {
			return NewLangError(TypeMismatch, rightType, leftType).At(b.Line, b.Column)
		}
		castExpr(b.Right, leftType)
		rightType = leftType
	}

	// If both are same type, result is that type
	if leftType != rightType && !canImplicitCast(rightType, leftType) {
		return NewLangError(TypeMismatch, rightType, leftType).At(b.Line, b.Column)
	}

	switch b.Operator {
	case SymbolLess, SymbolLessEqual, SymbolGreater, SymbolGreaterEqual, SymbolEqual, SymbolNotEqual:
		b.ReturnType = PrimitiveB1
		return nil
	default:
		b.ReturnType = leftType
		return nil
	}
}

func (tc *TypeChecker) AnalyzeCallExpr(c *CallExpr) error {
	// Resolve function symbol
	f, found := tc.GlobalScope.Resolve(c.Name)
	if !found {
		line, col := c.Pos()
		return NewLangError(InvalidFunctionCall, c.Name).At(line, col)
	}

	// Ensure it's a function (not a variable)
	fn, ok := f.(*Function)
	if !ok {
		line, col := c.Pos()
		return NewLangError(InvalidFunctionCall, c.Name).At(line, col)
	}

	// FIXME: Fix for "tuỳ" type parameter
	// Check argument count
	if len(c.Arguments) != len(fn.Parameters) {
		line, col := c.Pos()
		return NewLangError(ArgumentCountMismatch, len(c.Arguments), len(fn.Parameters), c.Name).At(line, col)
	}

	// Check argument types
	for i, arg := range c.Arguments {
		tc.AnalyzeExpression(arg)
		argType := tc.getExprType(arg)
		paramType := fn.Parameters[i].Type
		// FIXME:
		if argType != paramType && paramType != PrimitiveAny && isLiteral(arg) {
			if !canLiteralCast(argType, paramType) {
				line, col := arg.Pos()
				return NewLangError(TypeMismatch, argType, paramType).At(line, col)
			}
			argType = paramType
		}

		if argType != paramType && paramType != PrimitiveAny && !canImplicitCast(argType, paramType) {
			line, col := arg.Pos()
			return NewLangError(ArgumentTypeMismatch, argType, paramType).At(line, col)
		}

		if paramType != PrimitiveAny {
			castExpr(arg, paramType)
		}
	}

	c.ReturnType = fn.ReturnType
	return nil
}

// Helper
func (tc *TypeChecker) getExprType(expr Expression) string {
	switch e := expr.(type) {
	case *Identifier:
		v, found := tc.CurrentScope.Resolve(e.Name)
		if !found {
			line, col := e.Pos()
			panic(NewLangError(UndeclaredIdentifier, e.Name).At(line, col))
		}

		switch id := v.(type) {
		case *Variable:
			return id.Type
		default:
			line, col := e.Pos()
			panic(NewLangError(UndeclaredIdentifier, e.Name).At(line, col))
		}

		// Fallthrough if not found
	case *NumberLiteral:
		return e.Type
	case *BinaryExpr:
		return e.ReturnType
	case *CallExpr:
		return e.ReturnType
	case *ExplicitCast:
		return e.Type
	default:
		return "Unknown"
	}
}

func isLiteral(expr Expression) bool {
	switch e := expr.(type) {
	case *NumberLiteral:
		return true
	case *BinaryExpr:
		return isLiteral(e.Left) && isLiteral(e.Right)
	default:
		return false
	}
}

func canExplicitCast(fromType, toType string) bool {
	switch fromType {
	case PrimitiveR64, PrimitiveR32, PrimitiveZ64, PrimitiveZ32, PrimitiveN64, PrimitiveN32:
		switch toType {
		case PrimitiveR64, PrimitiveR32, PrimitiveZ64, PrimitiveZ32, PrimitiveN64, PrimitiveN32:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func canLiteralCast(fromType, toType string) bool {
	switch fromType {
	case PrimitiveR64:
		switch toType {
		case PrimitiveR64, PrimitiveR32:
			return true
		default:
			return false
		}
	case PrimitiveZ64:
		switch toType {
		case PrimitiveZ64, PrimitiveZ32, PrimitiveN64, PrimitiveN32:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

// Secretly cast to widen a type if possible
func canImplicitCast(fromType, toType string) bool {
	if fromType == toType {
		return true
	}

	switch fromType {
	case PrimitiveR64:
		switch toType {
		case PrimitiveR64: // Safe
			return true
		default:
			return false
		}
	case PrimitiveR32:
		switch toType {
		case PrimitiveR32, PrimitiveR64:
			return true
		default:
			return false
		}
	case PrimitiveZ64:
		switch toType {
		case PrimitiveZ64:
			return true
		default:
			return false
		}
	case PrimitiveZ32:
		switch toType {
		case PrimitiveZ32, PrimitiveZ64:
			return true
		default:
			return false
		}
	case PrimitiveN64:
		switch toType {
		case PrimitiveN64:
			return true
		default:
			return false
		}
	case PrimitiveN32:
		switch toType {
		case PrimitiveN32, PrimitiveN64:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func castExpr(expr Expression, toType string) {
	switch e := expr.(type) {
	case *Identifier:
		e.Type = toType
	case *NumberLiteral:
		e.Type = toType
	case *BinaryExpr:
		castExpr(e.Left, toType)
		castExpr(e.Right, toType)
		e.ReturnType = toType
	case *CallExpr:
		e.ReturnType = toType
	default:
	}
}

// Turns floating point numbers like "1.0" to "1"
func truncanteFloatString(s string) string {
	if dot := strings.Index(s, "."); dot != -1 {
		return s[:dot]
	}
	return s
}
