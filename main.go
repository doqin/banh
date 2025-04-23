package main

import (
	"fmt"
	"log"
	"os"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/types"
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
    log.Fatal("Error reading",err)
  } 
  fmt.Print(string(data))
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
  defer f.Close()
  f.WriteString(m.String())
}
