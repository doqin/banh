package main

import (
	"fmt"
	"reflect"
)

// TODO: Handle default values

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
		return fmt.Errorf("câu lệnh khai báo không xác định")
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
		Parameters: []*Variable{{Name: "giá_trị", Type: &PrimitiveType{Name: PrimitiveAny}}},
		ReturnType: &PrimitiveType{Name: PrimitiveZ32},
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

// FIXME: Fix the binary expression casting error

func (tc *TypeChecker) AnalyzeStatement(stmt Statement, expectedReturnType Type) error {
	switch s := stmt.(type) {
	case *VarDecl:
		err := tc.AnalyzeExpression(s.Value)
		if err != nil {
			return err
		}
		err = tc.AnalyzeType(s.Var.Type, s.Value)
		if err != nil {
			return err
		}
		// Declare variable
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
		err = tc.AnalyzeType(expectedReturnType, s.Value)
		if err != nil {
			return err
		}
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
		return fmt.Errorf("câu lệnh không xác định")
	}
}

func (tc *TypeChecker) AnalyzeType(checker Type, checked Expression) error {
	checkedType := tc.getExprType(checked)
	// If is same type then ok
	if isSameTypeAndName(checker, checkedType) {
		return nil
	}

	chcker, ok1 := checker.(*ContainerType)
	_, ok2 := checkedType.(*ContainerType)

	// If both are containers
	if ok1 && ok2 {
		checked, ok := checked.(*ArrayLiteral)
		// If expression is not of array literal then you can't implicit cast
		if !ok {
			line, col := checked.Pos()
			return NewLangError(TypeMismatch, checkedType.String(), checker.String()).At(line, col)
		}
		// Check type and cast for each elements
		for _, elem := range checked.Elements {
			err := tc.AnalyzeType(chcker.ElementType, elem)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// If one of them is container
	if ok1 || ok2 {
		line, col := checked.Pos()
		return NewLangError(TypeMismatch, checkedType.String(), chcker.String()).At(line, col)
	}

	// Handle if initializer is a literal
	if isLiteral(checked) {
		if !canLiteralCast(checkedType, checker) {
			line, col := checked.Pos()
			return NewLangError(TypeMismatch, checkedType.String(), checker.String()).At(line, col)
		}
	} else if !canImplicitCast(checkedType, checker) { // Handle if a type can be widen
		line, col := checked.Pos()
		return NewLangError(TypeMismatch, checkedType.String(), checker.String()).At(line, col)
	}

	// Finally cast after safty causions
	err := castExpr(checked, checker)
	if err != nil {
		return err
	}
	return nil
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
	case *ArrayLiteral:
		err := tc.AnalyzeArrayLiteral(e)
		if err != nil {
			return err
		}
		return nil
	case *IndexExpr:
		err := tc.AnalyzeIndexExpr(e)
		if err != nil {
			return err
		}
		return nil
	default:
		line, col := e.Pos()
		return NewLangError(UnknownExpression).At(line, col)
	}
}

func (tc *TypeChecker) AnalyzeArrayLiteral(a *ArrayLiteral) error {
	// TODO: Implement it
	if len(a.Elements) == 0 {
		a.Type = &ContainerType{Kind: ContainerArray, ElementType: &PrimitiveType{Name: PrimitiveAny}}
		return nil
	}

	var inferredType Type
	for i, elem := range a.Elements {
		err := tc.AnalyzeExpression(elem)
		if err != nil {
			return err
		}

		elemType := tc.getExprType(elem)
		if i == 0 {
			inferredType = elemType
			continue
		}

		if isLiteral(elem) && canLiteralCast(elemType, inferredType) {
			err := castExpr(elem, inferredType)
			if err != nil {
				return err
			}
			continue
		}

		if !isSameTypeAndName(elemType, inferredType) && !canImplicitCast(elemType, inferredType) {
			line, col := elem.Pos()
			return NewLangError(TypeMismatch, elemType.String(), inferredType.String()).At(line, col)
		}
	}

	a.Type = &ContainerType{
		Kind:        ContainerArray,
		ElementType: inferredType,
		Bounds: []Expression{
			&NumberLiteral{Value: "0", Type: PrimitiveType{Name: PrimitiveZ64}},
			&NumberLiteral{Value: fmt.Sprintf("%d", len(a.Elements)-1), Type: PrimitiveType{Name: PrimitiveZ64}}}}
	return nil
}

func (tc *TypeChecker) AnalyzeExplicitCast(e *ExplicitCast) error {
	castType := e.Type
	err := tc.AnalyzeExpression(e.Argument)
	if err != nil {
		return err
	}
	exprType := tc.getExprType(e.Argument)
	if !canExplicitCast(exprType, &castType) {
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
		// FIXME: Handle function name clashing with variable name
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
	leftTyp, ok1 := leftType.(*PrimitiveType)
	rightTyp, ok2 := rightType.(*PrimitiveType)
	if !ok1 || !ok2 {
		return NewLangError(ExpectToken, "kiểu dữ liệu nguyên thuỷ").At(b.Line, b.Column)
	}

	if b.Operator == KeywordVa || b.Operator == KeywordHoac {

		if leftTyp.Name != PrimitiveB1 {
			return NewLangError(ErrorBinaryExpr, leftTyp.Name, PrimitiveB1).At(b.Line, b.Column)
		}
		if rightTyp.Name != PrimitiveB1 {
			return NewLangError(ErrorBinaryExpr, rightTyp.Name, PrimitiveB1).At(b.Line, b.Column)
		}
		b.ReturnType.Name = PrimitiveB1
		return nil
	}

	if leftTyp.Name == rightTyp.Name {
		b.ReturnType.Name = leftTyp.Name
		return nil
	}

	if isLiteral(b.Left) && !isLiteral(b.Right) {
		if !canLiteralCast(leftType, rightType) {
			return NewLangError(ErrorBinaryExpr, leftType, rightType).At(b.Line, b.Column)
		}
		castExpr(b.Left, rightType)
		leftTyp.Name = rightTyp.Name
	} else if !isLiteral(b.Left) && isLiteral(b.Right) {
		if !canLiteralCast(rightType, leftType) {
			return NewLangError(ErrorBinaryExpr, rightType, leftType).At(b.Line, b.Column)
		}
		castExpr(b.Right, leftType)
		rightTyp.Name = leftTyp.Name
	}

	if leftType != rightType && !canImplicitCast(rightType, leftType) {
		return NewLangError(ErrorBinaryExpr, rightType, leftType).At(b.Line, b.Column)
	}

	switch b.Operator {
	case SymbolLess, SymbolLessEqual, SymbolGreater, SymbolGreaterEqual, SymbolEqual, SymbolNotEqual:
		b.ReturnType.Name = PrimitiveB1
		return nil
	default:
		b.ReturnType.Name = leftTyp.Name
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
		err := tc.AnalyzeExpression(arg)
		if err != nil {
			return err
		}
		argType := tc.getExprType(arg)
		paramType := fn.Parameters[i].Type
		if isSameTypeAndName(argType, paramType) || paramType.String() == PrimitiveAny {
			continue
		}

		if isLiteral(arg) {
			if !canLiteralCast(argType, paramType) {
				line, col := arg.Pos()
				return NewLangError(ArgumentTypeMismatch, argType, paramType).At(line, col)
			}
			argType = paramType
		}

		if !canImplicitCast(argType, paramType) {
			line, col := arg.Pos()
			return NewLangError(ArgumentTypeMismatch, argType, paramType).At(line, col)
		}

		castExpr(arg, paramType)
	}

	c.ReturnType = fn.ReturnType
	return nil
}

func (tc *TypeChecker) AnalyzeIndexExpr(i *IndexExpr) error {
	containerType := ContainerType{}

	// Check collection type
	switch collec := i.Collection.(type) {
	case *Identifier:
		err := tc.AnalyzeIdentifier(collec)
		if err != nil {
			return err
		}
		contain, ok := collec.Type.(*ContainerType)
		if !ok {
			line, col := i.Pos()
			return NewLangError(InvalidArrayAccessType).At(line, col)
		}
		containerType = *contain
	case *IndexExpr:
		err := tc.AnalyzeIndexExpr(collec) // I think it's okay?
		if err != nil {
			return err
		}
		typ := tc.getExprType(collec)
		contain, ok := typ.(*ContainerType)
		if !ok {
			line, col := i.Pos()
			return NewLangError(InvalidArrayAccessType).At(line, col)
		}
		containerType = *contain
	default:
		line, col := i.Pos()
		return NewLangError(InvalidArrayAccessType).At(line, col)
	}
	// Check indexing dimension
	if len(i.Indices) != containerType.Dimensions {
		line, col := i.Pos()
		return NewLangError(InvalidArrayAccessDim, len(i.Indices), containerType.Dimensions).At(line, col)
	}

	// Check indexing type
	for _, index := range i.Indices {
		typ := tc.getExprType(index)
		if !isTypeNumber_Type(typ) {
			line, col := index.Pos()
			return NewLangError(InvalidArrayAccessIndex).At(line, col)
		}
	}
	return nil
}

// Helper
func (tc *TypeChecker) getExprType(expr Expression) Type {
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
		return &e.Type
	case *BinaryExpr:
		return &e.ReturnType
	case *CallExpr:
		return e.ReturnType
	case *ExplicitCast:
		return &e.Type
	case *ArrayLiteral:
		return e.Type
	case *IndexExpr:
		switch collec := e.Collection.(type) {
		case *Identifier:
			typ, ok := collec.Type.(*ContainerType)
			if !ok {
				line, col := e.Pos()
				panic(NewLangError(InvalidArrayAccessType).At(line, col))
			}
			return typ
		case *IndexExpr:
			typ := tc.getExprType(collec)
			containerType, ok := typ.(*ContainerType)
			if !ok {
				line, col := e.Pos()
				panic(NewLangError(InvalidArrayAccessType).At(line, col))
			}
			// Removes one nested indexing
			return containerType.ElementType
		default:
			line, col := e.Pos()
			panic(NewLangError(InvalidArrayAccessType).At(line, col))
		}
	default:
		return &UnknownType{Name: "Unknown"}
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

// Handles explicit casting of primitive types
func canExplicitCast(fromType, toType Type) bool {
	if isSameTypeAndName(fromType, toType) {
		return true
	}

	fromTyp, ok1 := fromType.(*PrimitiveType)
	toTyp, ok2 := toType.(*PrimitiveType)

	if !ok1 || !ok2 {
		return false
	}

	switch fromTyp.Name {
	case PrimitiveR64, PrimitiveR32, PrimitiveZ64, PrimitiveZ32, PrimitiveN64, PrimitiveN32:
		switch toTyp.Name {
		case PrimitiveR64, PrimitiveR32, PrimitiveZ64, PrimitiveZ32, PrimitiveN64, PrimitiveN32:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func canLiteralCast(fromType, toType Type) bool {
	fromTyp, ok1 := fromType.(*PrimitiveType)
	toTyp, ok2 := toType.(*PrimitiveType)

	if !ok1 || !ok2 {
		return false
	}

	switch fromTyp.Name {
	case PrimitiveR64:
		switch toTyp.Name {
		case PrimitiveR64, PrimitiveR32:
			return true
		default:
			return false
		}
	case PrimitiveZ64:
		switch toTyp.Name {
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
func canImplicitCast(fromType, toType Type) bool {
	if isSameTypeAndName(fromType, toType) {
		return true
	}

	fromTyp, ok1 := fromType.(*PrimitiveType)
	toTyp, ok2 := toType.(*PrimitiveType)

	if !ok1 || !ok2 {
		fromTyp, ok3 := fromType.(*ContainerType)
		toTyp, ok4 := toType.(*ContainerType)
		if !ok3 || !ok4 {
			return false
		}
		return canImplicitCast(fromTyp.ElementType, toTyp.ElementType)
	}
	fromTypName := fromTyp.Name
	toTypName := toTyp.Name

	switch fromTypName {
	case PrimitiveR64:
		switch toTypName {
		case PrimitiveR64: // Safe
			return true
		default:
			return false
		}
	case PrimitiveR32:
		switch toTypName {
		case PrimitiveR32, PrimitiveR64:
			return true
		default:
			return false
		}
	case PrimitiveZ64:
		switch toTypName {
		case PrimitiveZ64:
			return true
		default:
			return false
		}
	case PrimitiveZ32:
		switch toTypName {
		case PrimitiveZ32, PrimitiveZ64:
			return true
		default:
			return false
		}
	case PrimitiveN64:
		switch toTypName {
		case PrimitiveN64:
			return true
		default:
			return false
		}
	case PrimitiveN32:
		switch toTypName {
		case PrimitiveN32, PrimitiveN64:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func castExpr(expr Expression, toType Type) error {
	switch e := expr.(type) {
	case *Identifier:
		e.Type = toType
		return nil
	case *NumberLiteral:
		toTyp, ok := toType.(*PrimitiveType)
		if !ok || !isTypeNumber_Type(toTyp) {
			return NewLangError(InvalidCasting, e.Type, toType).At(e.Line, e.Column)
		}
		e.Type.Name = toTyp.Name
		return nil
	case *BinaryExpr:
		err := castExpr(e.Left, toType)
		if err != nil {
			return err
		}
		err = castExpr(e.Right, toType)
		if err != nil {
			return err
		}
		toTyp, ok := toType.(*PrimitiveType)
		if !ok || !isTypeNumber_Type(toTyp) {
			return NewLangError(InvalidCasting, e.ReturnType, toType).At(e.Line, e.Column)
		}
		e.ReturnType.Name = toTyp.Name
		return nil
	case *CallExpr:
		e.ReturnType = toType
		return nil
	case *ArrayLiteral:
		toType, ok := toType.(*ContainerType)
		if !ok {
			return NewLangError(InvalidCasting, e.Type, toType).At(e.Line, e.Column)
		}
		e.Type = toType
		for _, elem := range e.Elements {
			err := castExpr(elem, toType.ElementType)
			if err != nil {
				return err
			}
		}
		return nil
	case *ExplicitCast:
		// I think it should be fine? Since it's handled by semantic analysis already
		return nil
	default:
		line, col := expr.Pos()
		// fmt.Printf("bruh: %+v", expr)
		return NewLangError(InvalidCasting, getExprType(expr), toType).At(line, col)
	}
}

func isSameTypeAndName(a, b Type) bool {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}
	return a.String() == b.String()
}

/*
// Turns floating point numbers like "1.0" to "1"
func truncanteFloatString(s string) string {
	if dot := strings.Index(s, "."); dot != -1 {
		return s[:dot]
	}
	return s
}
*/
