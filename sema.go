package main

import (
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
func (s *Scope) Declare(name string, v any) {
	switch typ := v.(type) {
	case *Variable:
		if _, exists := s.Symbols[name]; exists {
			panic(NewLangError(RedeclarationVar, name).At(typ.Line, typ.Column))
		}
		s.Symbols[name] = typ
	case *Function:
		if _, exists := s.Functions[name]; exists {
			panic(NewLangError(RedeclarationFunction, name).At(typ.Line, typ.Column))
		}
		s.Functions[name] = typ
	default:
		panic("Câu lệnh khai báo không xác định")
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
func (tc *TypeChecker) AnalyzeProgram(p *Program) {
	tc.GlobalScope = NewScope(nil)

	// First, declare all functions (for forward reference)
	for _, fn := range p.Functions {
		tc.GlobalScope.Declare(fn.Name, &Function{
			Name:       fn.Name,
			Parameters: fn.Parameters,
			ReturnType: fn.ReturnType,
			Line:       fn.Line,
			Column:     fn.Column,
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
	case *IfStmt:
		tc.AnalyzeExpression(s.Condition)
		for _, stmt := range s.ThenBlock {
			tc.AnalyzeStatement(stmt, expectedReturnType)
		}
		if s.ElseBlock != nil {
			for _, stmt = range s.ElseBlock {
				tc.AnalyzeStatement(stmt, expectedReturnType)
			}
		}
	default:
		panic("Câu lệnh không xác định")
	}
}

func (tc *TypeChecker) AnalyzeExpression(expr Expression) {
	switch e := expr.(type) {
	case *Identifier:
		v, found := tc.CurrentScope.Resolve(e.Name)
		if !found {
			line, col := e.Pos()
			panic(NewLangError(UndeclaredIdentifier, e.Name).At(line, col))
		}

		switch typ := v.(type) {
		case *Variable:
			e.Type = typ.Type
		case *Function:
			//FIXME: Handle function name clashing with variable name
			// Rather important
			line, col := e.Pos()
			panic(NewLangError(InvalidIdentifierUsage, e.Name).At(line, col))
		default:
			line, col := e.Pos()
			panic(NewLangError(UnknownIdentifierType).At(line, col))
		}
	case *NumberLiteral:
		// Is already R64
	case *BinaryExpr:
		tc.AnalyzeExpression(e.Left)
		tc.AnalyzeExpression(e.Right)
		leftType := tc.getExprType(e.Left)
		rightType := tc.getExprType(e.Right)

		if e.Operator == KeywordVa || e.Operator == KeywordHoac {
			if leftType != PrimitiveB1 {
				panic(NewLangError(TypeMismatch, leftType, PrimitiveB1).At(e.Line, e.Column))
			}
			if rightType != PrimitiveB1 {
				panic(NewLangError(TypeMismatch, rightType, PrimitiveB1).At(e.Line, e.Column))
			}
			e.ReturnType = PrimitiveB1
			break
		}

		if leftType == rightType {
			e.ReturnType = leftType
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

		switch e.Operator {
			case SymbolLess, SymbolLessEqual, SymbolGreater, SymbolGreaterEqual, SymbolEqual, SymbolNotEqual:
				e.ReturnType = PrimitiveB1
			default:
				e.ReturnType = leftType
		}
	case *CallExpr:
		// Resolve function symbol
		f, found := tc.GlobalScope.Resolve(e.Name)
		if !found {
			line, col := e.Pos()
			panic(NewLangError(InvalidFunctionCall, e.Name).At(line, col))
		}

		// Ensure it's a function (not a variable)
		fn, ok := f.(*Function)

		if !ok {
			line, col := e.Pos()
			panic(NewLangError(InvalidFunctionCall, e.Name).At(line, col))
		}

		// Check argument count
		if len(e.Arguments) != len(fn.Parameters) {
			line, col := e.Pos()
			panic(NewLangError(ArgumentCountMismatch, len(e.Arguments), len(fn.Parameters), e.Name).At(line, col))
		}

		for i, arg := range e.Arguments {
			tc.AnalyzeExpression(arg)
			argType := tc.getExprType(arg)
			paramType := fn.Parameters[i].Type
			// FIXME:
			if argType != paramType && isLiteral(arg) {
				if !canLiteralCast(argType, paramType) {
					line, col := arg.Pos()
					panic(NewLangError(TypeMismatch, argType, paramType).At(line, col))
				}
				argType = paramType
			}

			if argType != paramType && !canImplicitCast(argType, paramType) {
				line, col := arg.Pos()
				panic(NewLangError(ArgumentTypeMismatch, argType, paramType).At(line, col))
			}
			castExpr(arg, paramType)
		}

		e.ReturnType = fn.ReturnType
	default:
		line, col := e.Pos()
		panic(NewLangError(UnknownExpression).At(line, col))
	}
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
