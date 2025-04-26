package main

import (
	"log"
	"os"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Fatal("Làm ơn xác định câu lệnh cho chương trình:\n\tnuong -> Tạo bản dựng cho chương trình\n\tan -> Chạy chương trình\n")
	}
	switch args[1] {
	case "nuong":
		nuong()
	case "an":
		an()
	case "hap":
		hap()
	default:
		log.Fatal("Không nhận diện được câu lệnh, làm ơn thử lại!")
	}
}
