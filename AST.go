package main

import (
	"github.com/llir/llvm/ir/value"
)

// TODO: Add unary expression + (container access + assignment)

// AST Node
type Node interface {
	Pos() (line, column int)
}

// Variable type
type Type interface {
	String() string
	IsPrimitive() bool
}

// Placeholder unknown
type UnknownType struct {
	Name string
}

func (u *UnknownType) String() string    { return u.Name }
func (u *UnknownType) IsPrimitive() bool { return false }

// Primitive type
type PrimitiveType struct {
	Name string
}

func (p *PrimitiveType) String() string    { return p.Name }
func (p *PrimitiveType) IsPrimitive() bool { return true }

type StructType struct {
	Name   string
	Fields map[string]Type
	Order  []string
}

func (s *StructType) String() string    { return s.Name }
func (s *StructType) IsPrimitive() bool { return false }

type ContainerType struct {
	Kind        string
	ElementType Type
	Dimensions  int
	Bounds      []Expression
}

func (c *ContainerType) String() string {
	str := c.Kind
	str += " E " + c.ElementType.String()
	return str
}
func (c *ContainerType) IsPrimitive() bool { return false }

type Program struct {
	Globals   []*Variable
	Functions []*Function
	Structs   []*StructDecl
}

type Function struct {
	Name       string
	Parameters []*Variable
	ReturnType Type
	Body       []Statement
	Line       int
	Column     int
}

func (f *Function) Pos() (int, int) { return f.Line, f.Column }

type StructDecl struct {
	Name   string
	Fields []*StructField
	Line   int
	Column int
}

func (s *StructDecl) Pos() (int, int) { return s.Line, s.Column }

type StructField struct {
	Name   string
	Type   Type
	Line   int
	Column int
}

func (s *StructField) Pos() (int, int) { return s.Line, s.Column }

type Variable struct {
	Name   string
	Type   Type
	Line   int
	Column int
}

func (v *Variable) Pos() (int, int) { return v.Line, v.Column }

type Statement interface {
	Node
	statementNode()
	Codegen(ctx *CodegenContext) (value.Value, error)
}

type Expression interface {
	Node
	expressionNode()
	Codegen(ctx *CodegenContext) (value.Value, error)
}

// Example statement
type VarDecl struct {
	Var    *Variable
	Value  Expression
	Line   int
	Column int
}

func (v *VarDecl) statementNode()  {}
func (v *VarDecl) Pos() (int, int) { return v.Line, v.Column }

type ReturnStmt struct {
	Value  Expression
	Line   int
	Column int
}

func (r *ReturnStmt) statementNode()  {}
func (r *ReturnStmt) Pos() (int, int) { return r.Line, r.Column }

type IfStmt struct {
	Condition Expression
	ThenBlock []Statement
	ElseBlock []Statement
	Line      int
	Column    int
}

func (i *IfStmt) statementNode()  {}
func (i *IfStmt) Pos() (int, int) { return i.Line, i.Column }

type RegExpr struct { // Is a statement
	Expr   Expression
	Line   int
	Column int
}

func (r *RegExpr) statementNode()  {}
func (r *RegExpr) Pos() (int, int) { return r.Line, r.Column }

// Example expressions
type Identifier struct {
	Name   string
	Type   Type
	Line   int
	Column int
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) Pos() (int, int) { return i.Line, i.Column }

type UninitializedExpr struct {
	Column int
	Line   int
}

func (u *UninitializedExpr) expressionNode() {}
func (u *UninitializedExpr) Pos() (int, int) { return u.Line, u.Column }

type NumberLiteral struct {
	Value  string
	Type   PrimitiveType
	Line   int
	Column int
}

func (n *NumberLiteral) expressionNode() {}
func (n *NumberLiteral) Pos() (int, int) { return n.Line, n.Column }

type StructLiteral struct {
	StructName string
	Fields     map[string]Expression
	Line       int
	Column     int
}

func (s *StructLiteral) expressionNode() {}
func (s *StructLiteral) Pos() (int, int) { return s.Line, s.Column }

type ArrayLiteral struct {
	Elements []Expression
	Type     Type
	Line     int
	Column   int
}

func (s *ArrayLiteral) expressionNode() {}
func (s *ArrayLiteral) Pos() (int, int) { return s.Line, s.Column }

type BinaryExpr struct {
	Left       Expression
	Operator   string
	Right      Expression
	ReturnType PrimitiveType
	Line       int
	Column     int
}

func (b *BinaryExpr) expressionNode() {}
func (b *BinaryExpr) Pos() (int, int) { return b.Line, b.Column }

type CallExpr struct {
	Name       string
	Arguments  []Expression
	ReturnType Type
	Line       int
	Column     int
}

func (c *CallExpr) expressionNode() {}
func (c *CallExpr) Pos() (int, int) { return c.Line, c.Column }

type ExplicitCast struct {
	Type     PrimitiveType // For now only allow primitive type casting
	Argument Expression
	Line     int
	Column   int
}

func (e *ExplicitCast) expressionNode() {}
func (e *ExplicitCast) Pos() (int, int) { return e.Line, e.Column }
