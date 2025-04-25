package main

import (
	"fmt"
	"log"
	"os"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
	"slices"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Fatal("No file to read from. Aborting!")
	}
	filename := args[1]
	fmt.Println("Reading: ", filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("Error reading", err)
	}
	content := string(data)
	// Debugging
	if slices.Contains(args, "--print-char") {
		runes := []rune(content)
		for index, letter := range runes {
			fmt.Printf("Character at byte %d: %c (U+%04X)\n", index, letter, letter)
		}
	}

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

	parser := NewParser(tokens)
	program, err := parser.ParseProgram()
	if err != nil {
		log.Fatal("Không thể parse chương trình:\n", err)
	}
	if slices.Contains(args, "--print-function") {
		printProgram(program)
	}
}

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

func test() {
	// Create a new LLVM IR module
	m := ir.NewModule()

	// Define the main function: int main()
	mainFunc := m.NewFunc("main", types.I32)

	// Create an entry basic block.
	entry := mainFunc.NewBlock("")

	// Return the constant integer 42.
	entry.NewRet(constant.NewInt(types.I32, 42))

	// Write the LLVM IR assembly to a file
	f, err := os.Create("main.ll")
	if err != nil {
		log.Fatal("Failed to create file:", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal("Couldn't close file:", err)
		}
	}(f)
	_, err = f.WriteString(m.String())
	if err != nil {
		return
	}
}
