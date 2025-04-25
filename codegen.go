package main

import (
	"fmt"
	"strconv"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

type CodegenContext struct {
	Module  *ir.Module
	Func    *ir.Func
	Block   *ir.Block
	Symbols map[string]value.Value
}

// Function gen
func (fn *Function) Codegen(ctx *CodegenContext) (*ir.Func, error) {
	// TODO: Add proper parameter and return type support

	// For simplicity, assume no parameters and return type i32
	var fnIR *ir.Func
	if fn.Name == "chính" {
		fnIR = ctx.Module.NewFunc("main", types.I32)
	} else {
		fnIR = ctx.Module.NewFunc(fn.Name, types.I32)
	}

	entry := fnIR.NewBlock("entry")
	ctx.Func = fnIR
	ctx.Block = entry
	ctx.Symbols = make(map[string]value.Value) // fresh scope

	for _, stmt := range fn.Body {
		_, err := stmt.Codegen(ctx)
		if err != nil {
			return nil, err
		}
	}

	// If no explicit return, add default return 0
	if !blockHasTerminator(ctx.Block) {
		ctx.Block.NewRet(constant.NewInt(types.I32, 0))
	}

	return fnIR, nil
}

func (v *VarDecl) Codegen(ctx *CodegenContext) (value.Value, error) {
	entryBlock := ctx.Func.Blocks[0]
	// Allocate space for variable in entry block
	alloca := entryBlock.NewAlloca(types.I32) // TODO: adjust type accordingly

	// Save the alloca in the symbol table
	ctx.Symbols[v.Var.Name] = alloca

	// Generate code for initializer expression if any
	if v.Value != nil {
		val, err := v.Value.Codegen(ctx)
		if err != nil {
			return nil, err
		}
		// Store the value into the allocated space
		ctx.Block.NewStore(val, alloca)
	}

	return alloca, nil
}

func (r *ReturnStmt) Codegen(ctx *CodegenContext) (value.Value, error) {
	val, err := r.Value.Codegen(ctx)
	if err != nil {
		return nil, err
	}
	ctx.Block.NewRet(val)
	return val, nil
}

func (id *Identifier) Codegen(ctx *CodegenContext) (value.Value, error) {
	alloca, ok := ctx.Symbols[id.Name]
	if !ok {
		// TODO: Handle this differently
		return nil, fmt.Errorf("unknown variable %s", id.Name)
	}
	return ctx.Block.NewLoad(types.I32, alloca), nil
}

func (n *NumberLiteral) Codegen(ctx *CodegenContext) (value.Value, error) {
	// TODO: Handle double and float and others later

	// Assuming integer literaels for now
	val, err := strconv.Atoi(n.Value)
	if err != nil {
		return nil, err
	}
	return constant.NewInt(types.I32, int64(val)), nil
}

func GenerateLLVMIR(prog *Program) (*ir.Module, error) {
	ctx := &CodegenContext{
		Module:  ir.NewModule(),
		Symbols: make(map[string]value.Value),
	}

	for _, fn := range prog.Functions {
		_, err := fn.Codegen(ctx)
		if err != nil {
			return nil, err
		}
	}

	return ctx.Module, nil
}

// Helper function
func blockHasTerminator(block *ir.Block) bool {
	if block.Term == nil {
		return false
	}
	return true
}
