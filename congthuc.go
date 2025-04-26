package main

type CongThuc struct {
	Goi      GoiConfig      `toml:"goi"`
	PhuThuoc PhuThuocConfig `toml:"phuthuoc"`
	BanDung  BanDungConfig  `toml:"bandung"`
}

type GoiConfig struct {
	Ten    string   `toml:"ten"`
	Ban    string   `toml:"ban"`
	TacGia []string `toml:"tacgia"`
}

type PhuThuocConfig struct {
}

type BanDungConfig struct {
	DiemVao string `toml:"diemvao"`
	Xuat    string `toml:"xuat"`
}
