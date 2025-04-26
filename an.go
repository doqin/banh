package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/BurntSushi/toml"
)

func an() {
	var cth CongThuc
	if _, err := toml.DecodeFile("congthuc.toml", &cth); err != nil {
		log.Fatal("Không thể tải được 'congthuc.toml':", err)
	}
	fmt.Println("🍽️ Đang ăn bánh...")
	xuat := cth.BanDung.Xuat
	cmd := exec.Command(xuat)
	out, err := cmd.CombinedOutput()
	fmt.Println(string(out))
	if err != nil {
		log.Fatal("Gặp sự cố chạy chương trình:\n", err)
	}
}
