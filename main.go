package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Fatal("No file to read from. Aborting!")
	}
	var files []string
	for _, arg := range args {
		if strings.HasSuffix(arg, ".bnh") {
			files = append(files, arg)
		}
	}
	fmt.Println("Đang đọc tệp: ", files[0])

	data, err := os.ReadFile(files[0])
	if err != nil {
		log.Fatal("Error reading", err)
	}
	content := string(data)
	// Debugging
	if slices.Contains(args, "--print-char") {
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
		if slices.Contains(args, "--print-token") {
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
	if slices.Contains(args, "--print-program") {
		printProgram(program)
		fmt.Println()
	}

	// Generate code
	module, err := GenerateLLVMIR(program)
	if err != nil {
		log.Fatal("Gặp sự cố khi tạo code:\n", err)
	}
	if slices.Contains(args, "--print-ir") {
		fmt.Println(module.String())
	}

	// Write '.ll' file
	fileName := files[0]
	output := fileName[:len(fileName)-len(filepath.Ext(fileName))]
	irFile, err := os.Create(output + ".ll")
	if err != nil {
		log.Fatal("Gặp sự cố khi tạo file IR:\n", err)
	}
	defer irFile.Close()
	irFile.Write([]byte(module.String()))

	// Generate executable
	cmd := exec.Command("llc", "-filetype=obj", "-o", output+".o", output+".ll")
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Gặp sự cố chạy lệnh 'llc': %v", err)
	}

	cmd = exec.Command("clang", output+".o", "-o", output)
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Gặp sự cố chạy lệnh 'clang': %v", err)
	}

	fmt.Printf("Chương trình được tạo ở: %s\n", output)
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
	default:
		fmt.Printf("%sUnknown Statement\n", indent)
	}
}

func printExpression(e Expression, indent string) {
	switch expr := e.(type) {
	case *Identifier:
		fmt.Printf("Identifier(%s)", expr.Name)
	case *NumberLiteral:
		fmt.Printf("NumberLiteral(%s: %s)", expr.Value, expr.Type)
	default:
		fmt.Printf("Unknown Expression")
	}
}
