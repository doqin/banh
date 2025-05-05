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

	module := compile(args, cth)

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
