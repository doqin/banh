package main

import (
	"fmt"
	"strconv"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

type CodegenContext struct {
	Module      *ir.Module
	Func        *ir.Func
	Block       *ir.Block
	Symbols     map[string]value.Value
	ifIDCounter int
}

// Function gen
func (fn *Function) Codegen(ctx *CodegenContext) (*ir.Func, error) {
	// Handle params
	params := make([]*ir.Param, len(fn.Parameters))
	for i, param := range fn.Parameters {
		paramType := llvmTypeFromPrimitive(param.Type)
		params[i] = ir.NewParam(param.Name, paramType)
	}
	var fnIR *ir.Func
	if fn.Name == "chính" {
		if fn.ReturnType != PrimitiveZ32 {
			return nil, NewLangError(ReturnTypeMismatch, fn.ReturnType, PrimitiveZ32).At(fn.Line, fn.Column)
		}
		fnIR = ctx.Module.NewFunc("main", llvmTypeFromPrimitive(fn.ReturnType), params...)
	} else {
		fnIR = ctx.Module.NewFunc(fn.Name, llvmTypeFromPrimitive(fn.ReturnType), params...)
	}

	entry := fnIR.NewBlock("entry")
	ctx.Func = fnIR
	ctx.Block = entry
	ctx.Symbols = make(map[string]value.Value) // fresh scope

	// Map function parameters to allocas and store initial value
	for i, param := range fn.Parameters {
		llvmParam := fnIR.Params[i]
		alloca := ctx.Block.NewAlloca(llvmParam.Type())
		ctx.Symbols[param.Name] = alloca
		ctx.Block.NewStore(llvmParam, alloca)
	}

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

func (r *RegExpr) Codegen(ctx *CodegenContext) (value.Value, error) {
	_, err := r.Expr.Codegen(ctx)
	if err != nil {
		return nil, err
	}
	return nil, nil
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

func (i *IfStmt) Codegen(ctx *CodegenContext) (value.Value, error) {
	condVal, err := i.Condition.Codegen(ctx)
	if err != nil {
		return nil, err
	}

	if condVal.Type() != types.I1 {
		var intType *types.IntType
		switch condVal.Type() {
		case types.I8:
			intType = types.I8
		case types.I16:
			intType = types.I16
		case types.I32:
			intType = types.I32
		case types.I64:
			intType = types.I64
		default:
			return nil, fmt.Errorf("unsupported condition type: %v", condVal.Type())
		}
		condVal = ctx.Block.NewICmp(enum.IPredNE, condVal, constant.NewInt(intType, 0))
	}

	ifID := ctx.NextIfID()
	thenBlock := ctx.Func.NewBlock(fmt.Sprintf("if.then.%d", ifID))
	elseBlock := ctx.Func.NewBlock(fmt.Sprintf("if.else.%d", ifID))
	leaveBlock := ctx.Func.NewBlock(fmt.Sprintf("if.end.%d", ifID))

	ctx.Block.NewCondBr(condVal, thenBlock, elseBlock)

	ctx.Block = thenBlock
	for _, stmt := range i.ThenBlock {
		_, err := stmt.Codegen(ctx)
		if err != nil {
			return nil, err
		}
	}
	if !blockHasTerminator(ctx.Block) {
		ctx.Block.NewBr(leaveBlock)
	}

	ctx.Block = elseBlock
	for _, stmt := range i.ElseBlock {
		_, err := stmt.Codegen(ctx)
		if err != nil {
			return nil, err
		}
	}
	if !blockHasTerminator(ctx.Block) {
		ctx.Block.NewBr(leaveBlock)
	}

	ctx.Block = leaveBlock
	return nil, nil

	// something something
}

func (id *Identifier) Codegen(ctx *CodegenContext) (value.Value, error) {
	alloca, ok := ctx.Symbols[id.Name]
	if !ok {
		// TODO: Handle this differently
		return nil, fmt.Errorf("unknown variable %s", id.Name)
	}
	return ctx.Block.NewLoad(alloca), nil
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
		if (leftVal.Type().Equal(types.Double) || leftVal.Type().Equal(types.Float)) &&
			(rightVal.Type().Equal(types.Double) || rightVal.Type().Equal(types.Float)) {
			return ctx.Block.NewFAdd(leftVal, rightVal), nil
		}
		return ctx.Block.NewAdd(leftVal, rightVal), nil
	case SymbolMinus:
		if (leftVal.Type().Equal(types.Double) || leftVal.Type().Equal(types.Float)) &&
			(rightVal.Type().Equal(types.Double) || rightVal.Type().Equal(types.Float)) {
			return ctx.Block.NewFSub(leftVal, rightVal), nil
		}
		return ctx.Block.NewSub(leftVal, rightVal), nil
	case SymbolAsterisk:
		if (leftVal.Type().Equal(types.Double) || leftVal.Type().Equal(types.Float)) &&
			(rightVal.Type().Equal(types.Double) || rightVal.Type().Equal(types.Float)) {
			return ctx.Block.NewFMul(leftVal, rightVal), nil
		}
		return ctx.Block.NewMul(leftVal, rightVal), nil
	case SymbolSlash:
		// For now treat unsigned as signed
		if leftVal.Type().Equal(types.I32) && rightVal.Type().Equal(types.I32) || leftVal.Type().Equal(types.I64) && rightVal.Type().Equal(types.I64) {
			return ctx.Block.NewSDiv(leftVal, rightVal), nil
		}
		if (leftVal.Type().Equal(types.Double) || leftVal.Type().Equal(types.Float)) && (rightVal.Type().Equal(types.Double) || rightVal.Type().Equal(types.Float)) {
			return ctx.Block.NewFDiv(leftVal, rightVal), nil
		}
		return nil, fmt.Errorf("Gặp sự cố khi thực hiện phép toán")
	case SymbolLess:
		if canICmp(leftVal, rightVal) {
			return ctx.Block.NewICmp(enum.IPredSLT, leftVal, rightVal), nil
		} else if canFCmp(leftVal, rightVal) {
			return ctx.Block.NewFCmp(enum.FPredOLT, leftVal, rightVal), nil
		}
		return nil, fmt.Errorf("Gặp sự cố khi thực hiện phép so sánh giữa '%v' và '%v'", leftVal.Type(), rightVal.Type())
	case SymbolLessEqual:
		if canICmp(leftVal, rightVal) {
			return ctx.Block.NewICmp(enum.IPredSLE, leftVal, rightVal), nil
		} else if canFCmp(leftVal, rightVal) {
			return ctx.Block.NewFCmp(enum.FPredOLE, leftVal, rightVal), nil
		}
		return nil, fmt.Errorf("Gặp sự cố khi thực hiện phép so sánh giữa '%v' và '%v'", leftVal.Type(), rightVal.Type())
	case SymbolGreater:
		if canICmp(leftVal, rightVal) {
			return ctx.Block.NewICmp(enum.IPredSGT, leftVal, rightVal), nil
		} else if canFCmp(leftVal, rightVal) {
			return ctx.Block.NewFCmp(enum.FPredOGT, leftVal, rightVal), nil
		}
		return nil, fmt.Errorf("Gặp sự cố khi thực hiện phép so sánh giữa '%v' và '%v'", leftVal.Type(), rightVal.Type())
	case SymbolGreaterEqual:
		if canICmp(leftVal, rightVal) {
			return ctx.Block.NewICmp(enum.IPredSGE, leftVal, rightVal), nil
		} else if canFCmp(leftVal, rightVal) {
			return ctx.Block.NewFCmp(enum.FPredOGE, leftVal, rightVal), nil
		}
		return nil, fmt.Errorf("Gặp sự cố khi thực hiện phép so sánh giữa '%v' và '%v'", leftVal.Type(), rightVal.Type())
	case SymbolEqual:
		if canICmp(leftVal, rightVal) {
			return ctx.Block.NewICmp(enum.IPredEQ, leftVal, rightVal), nil
		} else if canFCmp(leftVal, rightVal) {
			return ctx.Block.NewFCmp(enum.FPredOEQ, leftVal, rightVal), nil
		}
		return nil, fmt.Errorf("Gặp sự cố khi thực hiện phép so sánh giữa '%v' và '%v'", leftVal.Type(), rightVal.Type())
	case SymbolNotEqual:
		if canICmp(leftVal, rightVal) {
			return ctx.Block.NewICmp(enum.IPredNE, leftVal, rightVal), nil
		} else if canFCmp(leftVal, rightVal) {
			return ctx.Block.NewFCmp(enum.FPredONE, leftVal, rightVal), nil
		}
		return nil, fmt.Errorf("Gặp sự cố khi thực hiện phép so sánh giữa '%v' và '%v'", leftVal.Type(), rightVal.Type())
	case KeywordVa:
		return ctx.Block.NewAnd(leftVal, rightVal), nil
	case KeywordHoac:
		return ctx.Block.NewOr(leftVal, rightVal), nil
	default:
		return nil, fmt.Errorf("Gặp sự cố khi thực hiện phép toán")
	}
}

func (c *CallExpr) Codegen(ctx *CodegenContext) (value.Value, error) {
	if c.Name == "in" {
		if len(c.Arguments) != 1 {
			return nil, fmt.Errorf("in() cần ít nhất một đối số")
		}
		argVal, err := c.Arguments[0].Codegen(ctx)
		if err != nil {
			return nil, err
		}
		printf := ctx.GetOrDeclarePrintf()

		var fmtStr string
		switch argVal.Type().String() {
		case "i32", "i64":
			fmtStr = "%d"
		case "float", "double":
			fmtStr = "%f"
		case "i8*": // string pointer
			fmtStr = "%s"
		default:
			return nil, fmt.Errorf("unsupported type for in(): %s", argVal.Type().String())
		}
		globalStr := ctx.GetOrCreateGlobalString(fmt.Sprintf("fmtstr_print_%s", fmtStr), fmtStr+"\n\x00")

		zero := constant.NewInt(types.I64, 0)
		fmtPtr := ctx.Block.NewGetElementPtr(globalStr, zero, zero) // I think?

		return ctx.Block.NewCall(printf, fmtPtr, argVal), nil
	}

	// Normal function calls
	callee := findFunction(ctx.Module, c.Name)
	if callee == nil {
		return nil, NewLangError(InvalidFunctionCall, c.Name)
	}

	// Check that the number of arguments matches the function's signature
	if len(c.Arguments) != len(callee.Params) {
		line, col := c.Pos()
		return nil, NewLangError(ArgumentCountMismatch, len(c.Arguments), len(callee.Params), callee.Name).At(line, col)
	}

	llvmArgs := make([]value.Value, len(c.Arguments))
	for i, arg := range c.Arguments {
		argVal, err := arg.Codegen(ctx)
		if err != nil {
			return nil, err
		}
		llvmArgs[i] = argVal
	}
	return ctx.Block.NewCall(callee, llvmArgs...), nil
}

func (e *ExplicitCast) Codegen(ctx *CodegenContext) (value.Value, error) {
	val, err := e.Argument.Codegen(ctx)
	if err != nil {
		return nil, err
	}

	targetType := llvmTypeFromPrimitive(e.Type)

	// Integer to Integer
	if srcInt, ok1 := val.Type().(*types.IntType); ok1 {
		if dstInt, ok2 := targetType.(*types.IntType); ok2 {
			if srcInt.BitSize < dstInt.BitSize {
				// Extend
				if e.Type == PrimitiveZ32 || e.Type == PrimitiveZ64 {
					return ctx.Block.NewZExt(val, targetType), nil
				} else {
					return ctx.Block.NewSExt(val, targetType), nil
				}
			} else if srcInt.BitSize > dstInt.BitSize {
				// Truncate
				return ctx.Block.NewTrunc(val, targetType), nil
			} else {
				// Same size
				return val, nil
			}
		}
	}

	// Integer to Float
	if _, ok1 := val.Type().(*types.IntType); ok1 {
		if _, ok2 := targetType.(*types.FloatType); ok2 {
			if e.Type == PrimitiveR32 || e.Type == PrimitiveR64 {
				// Signed int to float
				return ctx.Block.NewSIToFP(val, targetType), nil
			}
		}
	}

	// Float to Integer
	if _, ok1 := val.Type().(*types.FloatType); ok1 {
		if _, ok2 := targetType.(*types.IntType); ok2 {
			// Float to signed int
			return ctx.Block.NewFPToSI(val, targetType), nil
		}
	}

	srcFloat, ok1 := val.Type().(*types.FloatType)
	dstFloat, ok2 := targetType.(*types.FloatType)
	// Float to Float
	if ok1 && ok2 {
		if srcFloat.Kind == types.FloatKindFloat && dstFloat.Kind == types.FloatKindDouble {
			// Float to Double
			return ctx.Block.NewFPExt(val, targetType), nil
		} else if srcFloat.Kind == types.FloatKindDouble && dstFloat.Kind == types.FloatKindFloat {
			// Double to Float
			return ctx.Block.NewFPTrunc(val, targetType), nil
		} else {
			// Same Type
			return val, nil
		}
	}

	return nil, fmt.Errorf("unsupported cast from %v to %v", val.Type(), targetType)
}

func GenerateLLVMIR(prog *Program) (*ir.Module, error) {
	ctx := &CodegenContext{
		Module:      ir.NewModule(),
		Symbols:     make(map[string]value.Value),
		ifIDCounter: 0,
	}

	ctx.GetOrDeclarePrintf() // Declare printf
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
	case PrimitiveB1:
		return types.I1
	case PrimitiveN32, PrimitiveZ32:
		return types.I32
	case PrimitiveN64, PrimitiveZ64:
		return types.I64
	case PrimitiveR32:
		return types.Float
	case PrimitiveR64:
		return types.Double
	case PrimitiveVoid:
		return types.Void
	default:
		panic("Không xác định được kiểu dữ liệu")
	}
}

func findFunction(module *ir.Module, funcName string) *ir.Func {
	for _, fn := range module.Funcs {
		if fn.Name() == funcName {
			return fn
		}
	}
	return nil
}

func canICmp(left value.Value, right value.Value) bool {
	leftType := left.Type()
	rightType := right.Type()

	if leftType.Equal(rightType) {
		if leftType.Equal(types.I32) || leftType.Equal(types.I64) {
			return true
		}
	}
	return false
}

func canFCmp(left value.Value, right value.Value) bool {
	leftType := left.Type()
	rightType := right.Type()

	if leftType.Equal(rightType) {
		if leftType.Equal(types.Float) || leftType.Equal(types.Double) {
			return true
		}
	}
	return false
}

func (ctx *CodegenContext) NextIfID() int {
	ctx.ifIDCounter++
	return ctx.ifIDCounter
}

func (ctx *CodegenContext) GetOrCreateGlobalString(name, value string) *ir.Global {
	for _, g := range ctx.Module.Globals {
		if g.Name() == name {
			return g
		}
	}
	strConst := constant.NewCharArrayFromString(value)
	global := ir.NewGlobalDef(name, strConst)
	global.Linkage = enum.LinkagePrivate // huh??
	global.Immutable = true
	ctx.Module.Globals = append(ctx.Module.Globals, global)
	return global
}

func (ctx *CodegenContext) GetOrDeclarePrintf() *ir.Func {
	printf := findFunction(ctx.Module, "printf")
	if printf != nil {
		return printf
	}
	// printf: i32(i8*, ...)
	printf = ctx.Module.NewFunc("printf", types.I32, ir.NewParam("fmt", types.NewPointer(types.I8)))
	printf.Sig.Variadic = true
	printf.Linkage = enum.LinkageExternal // right? 🤔
	return printf
}
