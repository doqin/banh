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
	for {
		tok := lexer.NextToken()
		fmt.Printf("%+v\n", tok)
		if tok.Type == TokenEOF {
			break
		}
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
