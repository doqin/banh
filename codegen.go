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
		if fn.ReturnType != PrimitiveZ32 {
			return nil, NewLangError(ReturnTypeMismatch, fn.ReturnType, PrimitiveZ32).At(fn.Line, fn.Column)
		}
		fnIR = ctx.Module.NewFunc("main", llvmTypeFromPrimitive(fn.ReturnType))
	} else {
		fnIR = ctx.Module.NewFunc(fn.Name, llvmTypeFromPrimitive(fn.ReturnType))
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
		switch fnIR.Sig.RetType {
		case types.I32:
			ctx.Block.NewRet(constant.NewInt(types.I32, 0))
			break
		case types.I64:
			ctx.Block.NewRet(constant.NewInt(types.I64, 0))
			break
		case types.Double:
			ctx.Block.NewRet(constant.NewFloat(types.Double, 0))
			break
		case types.Void: // Is this necessary lol?
			ctx.Block.NewRet(nil)
			break
		}
	}
	return fnIR, nil
}

func (v *VarDecl) Codegen(ctx *CodegenContext) (value.Value, error) {
	entryBlock := ctx.Func.Blocks[0]
	// Allocate space for variable in entry block
	alloca := entryBlock.NewAlloca(llvmTypeFromPrimitive(v.Var.Type))

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
	return ctx.Block.NewLoad(llvmTypeFromPrimitive(id.Type), alloca), nil
}

func (n *NumberLiteral) Codegen(ctx *CodegenContext) (value.Value, error) {
	switch n.Type {
	case PrimitiveN32, PrimitiveZ32:
		val, err := strconv.ParseInt(n.Value, 10, 32)
		if err != nil {
			return nil, err
		}
		return constant.NewInt(types.I32, val), nil
	case PrimitiveN64, PrimitiveZ64:
		val, err := strconv.ParseInt(n.Value, 10, 64)
		if err != nil {
			return nil, err
		}
		return constant.NewInt(types.I64, val), nil
	case PrimitiveR32:
		val, err := strconv.ParseFloat(n.Value, 32)
		if err != nil {
			return nil, err
		}
		return constant.NewFloat(types.Float, val), nil
	case PrimitiveR64:
		val, err := strconv.ParseFloat(n.Value, 64)
		if err != nil {
			return nil, err
		}
		return constant.NewFloat(types.Double, val), nil
	default:
		panic("Không xác định được kiểu dữ liệu của số")
	}
}

func (b *BinaryExpr) Codegen(ctx *CodegenContext) (value.Value, error) {
	leftVal, err := b.Left.Codegen(ctx)
	if err != nil {
		return nil, err
	}
	rightVal, err := b.Right.Codegen(ctx)
	if err != nil {
		return nil, err
	}

	switch b.Operator {
	case SymbolPlus:
		return ctx.Block.NewAdd(leftVal, rightVal), nil
	case SymbolMinus:
		return ctx.Block.NewSub(leftVal, rightVal), nil
	case SymbolAsterisk:
		return ctx.Block.NewMul(leftVal, rightVal), nil
	case SymbolSlash:
		// For now treat unsigned as signed
		if leftVal.Type().Equal(types.I32) {
			return ctx.Block.NewSDiv(leftVal, rightVal), nil
		}
		if leftVal.Type().Equal(types.Double) || leftVal.Type().Equal(types.Float) {
			return ctx.Block.NewFAdd(leftVal, rightVal), nil
		}
		panic("Gặp sự cố khi thực hiện phép toán")
	default:
		panic("Gặp sự cố khi thực hiện phép toán")
	}
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

func llvmTypeFromPrimitive(name string) types.Type {
	switch name {
	case PrimitiveN32, PrimitiveZ32:
		return types.I32
	case PrimitiveN64, PrimitiveZ64:
		return types.I64
	case PrimitiveR32:
		return types.Float
	case PrimitiveR64:
		return types.Double
	default:
		panic("Không xác định được kiểu dữ liệu")
	}
}
