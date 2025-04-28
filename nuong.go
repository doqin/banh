package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"slices"

	"github.com/BurntSushi/toml"
)

func nuong() {
	// Reading the 'congthuc.toml' file
	var cth CongThuc
	if _, err := toml.DecodeFile("congthuc.toml", &cth); err != nil {
		log.Fatal("Không thể tải 'congthuc.toml':", err)
	}
	args := os.Args
	if slices.Contains(args, "--in-cong-thuc") {
		fmt.Println("📦 Gói:", cth.Goi.Ten)
		fmt.Println("🔖 Phiên bản:", cth.Goi.Ban)
		fmt.Println("✍️ Tác giả:", cth.Goi.TacGia)
		fmt.Println("📂 Điểm vào:", cth.BanDung.DiemVao)
		fmt.Println("📦 Xuất ra:", cth.BanDung.Xuat)
	}

	fmt.Println("🔥 Đang nướng bánh...")

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
	checker.AnalyzeProgram(program)

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

	// Write '.ll' file
	output := cth.BanDung.Xuat
	irFile, err := os.Create(output + ".ll")
	if err != nil {
		log.Fatal("Gặp sự cố khi tạo file IR:\n", err)
	}
	defer irFile.Close()
	irFile.Write([]byte(module.String()))

	// Generate executable
	cmd := exec.Command("llc", "-filetype=obj", "-o", output+".o", output+".ll")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		fmt.Printf("Gặp sự cố chạy lệnh 'llc': %v\n", err)
		os.Exit(1)
	}

	cmd = exec.Command("clang", output+".o", "-o", output)
	out, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		log.Fatalf("Gặp sự cố chạy lệnh 'clang': %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Bánh đã chín! Có thể ăn được rồi.")
}

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
		fmt.Printf("      - %s: %s (Line %d, Column %d)\n", param.Name, param.Type, param.Line, param.Column)
	}
	fmt.Printf("    Return Type: %s\n", f.ReturnType)
	fmt.Printf("    Body:\n")
	for _, stmt := range f.Body {
		printStatement(stmt, "      ")
	}
	fmt.Println("")
}

func printStatement(s Statement, indent string) {
	switch stmt := s.(type) {
	case *VarDecl:
		fmt.Printf("%sVarDecl: %s: %s = ", indent, stmt.Var.Name, stmt.Var.Type)
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
	default:
		fmt.Printf("%sUnknown Statement\n", indent)
	}
}

func printExpression(e Expression, indent string) {
	switch expr := e.(type) {
	case *Identifier:
		fmt.Printf("Identifier(%s: %s)", expr.Name, expr.Type)
	case *NumberLiteral:
		fmt.Printf("NumberLiteral(%s: %s)", expr.Value, expr.Type)
	case *BinaryExpr:
		fmt.Print("BinaryExpr(")
		printExpression(expr.Left, indent)
		fmt.Print(" ", expr.Operator, " ")
		printExpression(expr.Right, indent)
		fmt.Print(")")
	case *CallExpr:
		fmt.Printf("CallExpr: %s(", expr.Name)
		for i, argument := range expr.Arguments {
			printExpression(argument, indent)
			if i < len(expr.Arguments)-1 {
				fmt.Print(", ")
			}
		}
		fmt.Printf(") -> %s", expr.ReturnType)
	default:
		fmt.Printf("Unknown Expression")
	}
}
