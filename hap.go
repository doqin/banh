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

	module := compile(args, cth)

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

	fmt.Println(string(output))
	os.Remove(tmpfile.Name())
}
