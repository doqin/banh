package main

import (
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/llir/llvm/ir"
)

// Helper functions
func printProgram(p *Program) {
	fmt.Println("Program:")
	for _, fn := range p.Functions {
		printFunction(fn)
	}
}

func printFunction(f *Function) {
	fmt.Printf("  Function: %s (Line %d, Column %d)\n", f.Name, f.Line, f.Column)
	fmt.Printf("    Parameters:\n")
	for _, param := range f.Parameters {
		fmt.Printf("      - %s: %s (Line %d, Column %d)\n", param.Name, param.Type.String(), param.Line, param.Column)
	}
	fmt.Printf("    Return Type: %s\n", f.ReturnType.String())
	fmt.Printf("    Body:\n")
	for _, stmt := range f.Body {
		printStatement(stmt, "      ")
	}
	fmt.Println("")
}

func printStatement(s Statement, indent string) {
	switch stmt := s.(type) {
	case *VarDecl:
		fmt.Printf("%sVarDecl: %s: %s = ", indent, stmt.Var.Name, stmt.Var.Type.String())
		printExpression(stmt.Value, "")
		fmt.Printf(" (Line %d, Column %d)\n", stmt.Line, stmt.Column)
	case *ReturnStmt:
		fmt.Printf("%sReturn: ", indent)
		printExpression(stmt.Value, "")
		fmt.Printf(" (Line %d, Column %d)\n", stmt.Line, stmt.Column)
	case *IfStmt:
		fmt.Print(indent, "IfStmt:\n", indent+"   ", "Condition: ")
		printExpression(stmt.Condition, indent+"   ")
		fmt.Printf(" (Line %d, Column %d)\n", stmt.Line, stmt.Column)
		fmt.Print(indent+"   ", "ThenBlock:\n")
		for _, stmt := range stmt.ThenBlock {
			printStatement(stmt, indent+"      ")
		}
		if stmt.ElseBlock != nil {
			fmt.Print(indent+"   ", "ElseBlock:\n")
			for _, stmt := range stmt.ElseBlock {
				printStatement(stmt, indent+"      ")
			}
		}
		fmt.Println("")
	case *RegExpr:
		fmt.Printf("%sRegExpr: ", indent)
		printExpression(stmt.Expr, "")
		fmt.Printf(" (Line %d, Column %d)\n", stmt.Line, stmt.Column)
	default:
		fmt.Printf("%sUnknown Statement\n", indent)
	}
}

func printExpression(e Expression, indent string) {
	switch expr := e.(type) {
	case *Identifier:
		fmt.Printf("Identifier(%s: %s)", expr.Name, expr.Type.String())
	case *NumberLiteral:
		fmt.Printf("NumberLiteral(%s: %s)", expr.Value, expr.Type.String())
	case *BinaryExpr:
		fmt.Print("BinaryExpr(\n")
		fmt.Printf("%s         ", indent)
		printExpression(expr.Left, indent+"   ")
		fmt.Printf("\n%s         ", indent)
		fmt.Print("Operator: ", expr.Operator, "\n")
		fmt.Printf("%s         ", indent)
		printExpression(expr.Right, indent+"   ")
		fmt.Printf("\n%s         )", indent)
	case *CallExpr:
		fmt.Printf("CallExpr: %s(", expr.Name)
		for i, argument := range expr.Arguments {
			printExpression(argument, indent)
			if i < len(expr.Arguments)-1 {
				fmt.Print(", ")
			}
		}
		fmt.Printf(") -> %s", expr.ReturnType.String())
	case *ArrayLiteral:
		fmt.Printf("ArrayLiteral: %s {\n", expr.Type.String())
		for i, elem := range expr.Elements {
			fmt.Printf("%s      ", indent)
			printExpression(elem, indent)
			if i+1 < len(expr.Elements) {
				fmt.Print(",\n")
			}
		}
		fmt.Printf("\n%s      }", indent)
	case *ExplicitCast:
		fmt.Printf("ExplicitCast: %s(", expr.Type.String())
		printExpression(expr.Argument, indent)
		fmt.Print(")")
	default:
		fmt.Printf("Unknown Expression")
	}
}

func compile(args []string, cth CongThuc) *ir.Module {
	data, err := os.ReadFile(cth.BanDung.DiemVao)
	if err != nil {
		log.Fatal("Error reading", err)
	}
	content := string(data)
	// Debugging
	if slices.Contains(args, "--in-ky-tu") {
		runes := []rune(content)
		for index, letter := range runes {
			fmt.Printf("Ký tự tại byte %d: %c (U+%04X)\n", index, letter, letter)
		}
	}

	// Lex source code
	lexer := NewLexer(content)
	var tokens []Token
	for {
		tok := lexer.NextToken()
		tokens = append(tokens, tok)
		if slices.Contains(args, "--in-token") {
			fmt.Printf("%+v\n", tok)
		}
		if tok.Type == TokenEOF {
			break
		}
	}

	// Parse tokens
	parser := NewParser(tokens)
	program, err := parser.ParseProgram()
	if err != nil {
		log.Fatal("Không thể parse chương trình:\n", err)
	}

	if slices.Contains(args, "--in-parse") {
		printProgram(program)
		fmt.Println()
	}

	checker := &TypeChecker{}
	err = checker.AnalyzeProgram(program)
	if err != nil {
		log.Fatal("Gặp sự cố kiểm tra chương trình:\n", err)
	}

	if slices.Contains(args, "--in-chuong-trinh") {
		printProgram(program)
		fmt.Println()
	}

	// Generate code
	module, err := GenerateLLVMIR(program)
	if err != nil {
		log.Fatal("Gặp sự cố khi tạo code:\n", err)
	}
	if slices.Contains(args, "--in-ir") {
		fmt.Println(module.String())
	}
	return module
}
