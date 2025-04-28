package main

const (
	KeywordHam      = "hàm"
	KeywordBien     = "biến"
	KeywordNeu      = "nếu"
	KeywordTrongKhi = "trong khi"
	KeywordKhongThi = "không thì"
	KeywordThi      = "thì"
	KeywordKetThuc  = "kết thúc"
	KeywordVa       = "và"
	KeywordHoac     = "hoặc"
	KeywordTraVe    = "trả về"
	KeywordThuTuc   = "thủ tục"
)

var Keywords = map[string]string{
	"hàm":  KeywordHam,
	"biến": KeywordBien,
	"nếu":  KeywordNeu,
	"và":   KeywordVa,
	"hoặc": KeywordHoac,
	"thì":  KeywordThi,
	// Multi-word keywords are handled in the lexer
}
