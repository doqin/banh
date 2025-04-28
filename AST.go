package main

import (

	"github.com/llir/llvm/ir/value"
)

type Node interface {
	Pos() (line, column int)
}

type Program struct {
	Functions []*Function
}

type Function struct {
	Name       string
	Parameters []*Variable
	ReturnType string
	Body       []Statement
	Line       int
	Column     int
}

func (f *Function) Pos() (int, int) { return f.Line, f.Column }

type Variable struct {
	Name   string
	Type   string
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
	GetType() string
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
	Expr Expression
	Line int
	Column int
}

func (r *RegExpr) statementNode() {}
func (r *RegExpr) Pos() (int, int) { return r.Line, r.Column}

// Example expressions
type Identifier struct {
	Name   string
	Type   string
	Line   int
	Column int
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) Pos() (int, int) { return i.Line, i.Column }
func (i *Identifier) GetType() string { return i.Type }

type NumberLiteral struct {
	Value  string
	Type   string
	Line   int
	Column int
}

func (n *NumberLiteral) expressionNode() {}
func (n *NumberLiteral) Pos() (int, int) { return n.Line, n.Column }
func (n *NumberLiteral) GetType() string { return n.Type }

type BinaryExpr struct {
	Left       Expression
	Operator   string
	Right      Expression
	ReturnType string
	Line       int
	Column     int
}

func (b *BinaryExpr) expressionNode() {}
func (b *BinaryExpr) Pos() (int, int) { return b.Line, b.Column }
func (b *BinaryExpr) GetType() string { return b.ReturnType }

type CallExpr struct {
	Name       string
	Arguments  []Expression
	ReturnType string
	Line       int
	Column     int
}

func (c *CallExpr) expressionNode() {}
func (c *CallExpr) Pos() (int, int) { return c.Line, c.Column }
func (c *CallExpr) GetType() string { return c.ReturnType }

type ExplicitCast struct {
	Type string
	Argument Expression
	Line int
	Column int
}

func (e *ExplicitCast) expressionNode() {}
func (e *ExplicitCast) Pos() (int, int) { return e.Line, e.Column}
func (e *ExplicitCast) GetType() string { return e.Type }
