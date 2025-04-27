package main

import (
	"strings"
)

// Scope represents a symbol table with optional parent scoping
type Scope struct {
	Parent  *Scope
	Symbols map[string]*Variable
}

// NewScope creates a new scope, optionally with a parent
func NewScope(parent *Scope) *Scope {
	return &Scope{
		Parent:  parent,
		Symbols: make(map[string]*Variable),
	}
}

// Declare adds a new variable to the current scope
// It panics if the name already exists in the current scope.
func (s *Scope) Declare(name string, v *Variable) {
	if _, exists := s.Symbols[name]; exists {
		panic(NewLangError(RedeclarationVar, name).At(v.Line, v.Column))
	}
	s.Symbols[name] = v
}

// Resolve looks for a variable in the current scope chain.
func (s *Scope) Resolve(name string) *Variable {
	if v, ok := s.Symbols[name]; ok {
		return v
	}
	if s.Parent != nil {
		return s.Parent.Resolve(name)
	}
	return nil // Not found
}

type TypeChecker struct {
	GlobalScope  *Scope
	CurrentScope *Scope
}

// Entry point
func (tc *TypeChecker) AnalyzeProgram(p *Program) {
	tc.GlobalScope = NewScope(nil)
	// First, declare all functions (for forward reference)
	for _, fn := range p.Functions {
		tc.GlobalScope.Declare(fn.Name, &Variable{
			Name:   fn.Name,
			Type:   "function",
			Line:   fn.Line,
			Column: fn.Column,
		})
	}

	// Then check function bodies
	for _, fn := range p.Functions {
		tc.AnalyzeFunction(fn)
	}
}

func (tc *TypeChecker) AnalyzeFunction(fn *Function) {
	tc.CurrentScope = NewScope(tc.GlobalScope)

	// Declare parameters
	for _, param := range fn.Parameters {
		tc.CurrentScope.Declare(param.Name, param)
	}

	// Analyze body statements
	for _, stmt := range fn.Body {
		tc.AnalyzeStatement(stmt, fn.ReturnType)
	}
}

func (tc *TypeChecker) AnalyzeStatement(stmt Statement, expectedReturnType string) {
	switch s := stmt.(type) {
	case *VarDecl:
		tc.AnalyzeExpression(s.Value)
		valType := tc.getExprType(s.Value)

		if valType != s.Var.Type && isLiteral(s.Value) {
			if !canLiteralCast(valType, s.Var.Type) {
				line, col := s.Pos()
				panic(NewLangError(TypeMismatch, valType, s.Var.Type).At(line, col))
			}
			valType = s.Var.Type
		}

		if valType != s.Var.Type && !canImplicitCast(valType, s.Var.Type) {
			line, col := s.Pos()
			panic(NewLangError(TypeMismatch, valType, s.Var.Type).At(line, col))
		}
		castExpr(s.Value, s.Var.Type)
		tc.CurrentScope.Declare(s.Var.Name, s.Var)
	case *ReturnStmt:
		tc.AnalyzeExpression(s.Value)
		valType := tc.getExprType(s.Value)

		if valType != expectedReturnType && isLiteral(s.Value) {
			if !canLiteralCast(valType, expectedReturnType) {
				line, col := s.Pos()
				panic(NewLangError(ReturnTypeMismatch, valType, expectedReturnType).At(line, col))
			}
			valType = expectedReturnType
		}

		if valType != expectedReturnType && !canImplicitCast(valType, expectedReturnType) {
			line, col := s.Pos()
			panic(NewLangError(ReturnTypeMismatch, valType, expectedReturnType).At(line, col))
		}
		castExpr(s.Value, expectedReturnType)
	default:
		panic("Câu lệnh không xác định")
	}
}

func (tc *TypeChecker) AnalyzeExpression(expr Expression) {
	switch e := expr.(type) {
	case *Identifier:
		v := tc.CurrentScope.Resolve(e.Name)
		if v == nil {
			line, col := e.Pos()
			panic(NewLangError(UndeclaredIdentifier, e.Name).At(line, col))
		}
		e.Type = v.Type
		// Ideally annotate the Identifier with type info

	case *NumberLiteral:
		// Is already R64
	case *BinaryExpr:
		tc.AnalyzeExpression(e.Left)
		tc.AnalyzeExpression(e.Right)
		leftType := tc.getExprType(e.Left)
		rightType := tc.getExprType(e.Right)

		if leftType == rightType {
			e.Type = leftType
			break
		}

		if isLiteral(e.Left) && !isLiteral(e.Right) {
			if !canLiteralCast(leftType, rightType) {
				panic(NewLangError(TypeMismatch, leftType, rightType).At(e.Line, e.Column))
			}
			castExpr(e.Left, rightType)
			leftType = rightType
		} else if !isLiteral(e.Left) && isLiteral(e.Right) {
			if !canLiteralCast(rightType, leftType) {
				panic(NewLangError(TypeMismatch, rightType, leftType).At(e.Line, e.Column))
			}
			castExpr(e.Right, leftType)
			rightType = leftType
		}

		// If both are same type, result is that type
		if leftType != rightType && !canImplicitCast(rightType, leftType) {
			panic(NewLangError(TypeMismatch, rightType, leftType).At(e.Line, e.Column))
		}
		e.Type = leftType
	default:
		panic("Biểu thức không xác định")
	}
}

// Helper
func (tc *TypeChecker) getExprType(expr Expression) string {
	switch e := expr.(type) {
	case *Identifier:
		v := tc.CurrentScope.Resolve(e.Name)
		if v != nil {
			return v.Type
		}
		// Fallthrough if not found
		return "Unknown" // Store resolved type later
	case *NumberLiteral:
		return e.Type
	case *BinaryExpr:
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
		e.Type = toType
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

func exprPrecedence(expr Expression) int {
	switch expr.(type) {
	case *BinaryExpr:
		return 10
	case *Identifier, *NumberLiteral:
		return 30
	default:
		return -1
	}
}
