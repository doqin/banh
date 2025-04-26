package main

import "strings"

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

		// If value is a NumberLiteral and needs adjustment
		if numLit, ok := s.Value.(*NumberLiteral); ok {
			if numLit.Type == PrimitiveR64 && (s.Var.Type == PrimitiveN32 || s.Var.Type == PrimitiveN64 || s.Var.Type == PrimitiveZ32 || s.Var.Type == PrimitiveZ64) {
				// "Cast" NumberLiteral by modifying its type and value
				newValue := truncanteFloatString(numLit.Value)
				numLit.Value = newValue
				numLit.Type = s.Var.Type
				valType = numLit.Type
			}
		}

		if !canAssign(valType, s.Var.Type) {
			line, col := s.Pos()
			panic(NewLangError(TypeMismatch, valType, expectedReturnType).At(line, col))
		}
		tc.CurrentScope.Declare(s.Var.Name, s.Var)
	case *ReturnStmt:
		tc.AnalyzeExpression(s.Value)
		valType := tc.getExprType(s.Value)
		if !canAssign(valType, expectedReturnType) {
			line, col := s.Pos()
			panic(NewLangError(ReturnTypeMismatch, valType, expectedReturnType).At(line, col))
		}
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
	default:
		return "Unknown"
	}
}

func canAssign(fromType, toType string) bool {
	if fromType == toType {
		return true
	}

	// Allow implicit casts from R64 to int types (N32, Z32, etc.)
	if fromType == PrimitiveR64 {
		switch toType {
		case PrimitiveN32, PrimitiveN64, PrimitiveZ32, PrimitiveZ64:
			return true
		}
	}
	// Add more implicit cast rules
	return false
}

// Turns floating point numbers like "1.0" to "1"
func truncanteFloatString(s string) string {
	if dot := strings.Index(s, "."); dot != -1 {
		return s[:dot]
	}
	return s
}
