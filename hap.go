package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/BurntSushi/toml"
)

func hap() {
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

	fmt.Println("🥟 Đang hấp bánh...")

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

	dir := filepath.Dir(cth.BanDung.Xuat)
	file := filepath.Base(cth.BanDung.Xuat)

	tmpfile, err := os.CreateTemp(dir, file+".ll")
	if err != nil {
		log.Fatal("Gặp sự cố khi tạo file tạm thời:", err)
	}
	defer tmpfile.Close()

	_, err = tmpfile.WriteString(module.String()) // Dump LLIR as string
	if err != nil {
		os.Remove(tmpfile.Name())
		log.Fatal("Gặp sự cố khi viết file tạm thời:", err)
	}

	// Run `lli` with the temporary file
	cmd := exec.Command("lli", tmpfile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		os.Remove(tmpfile.Name())
		log.Fatalf("Gặp sự cố khi chạy 'lli':\n %v\nXuất: %s", err, output)
	}

	log.Println(string(output))
	os.Remove(tmpfile.Name())
}
