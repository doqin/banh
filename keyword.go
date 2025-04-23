package main

const (
	KeywordHam            = "hàm"
	KeywordBien           = "biến"
	KeywordNeu            = "nếu"
	KeywordTrongKhi       = "trong khi"
	KeywordLonHon         = "lớn hơn"
	KeywordNhoHon         = "nhỏ hơn"
	KeywordLonHonHoacBang = "lớn hơn hoặc bằng"
	KeywordNhoHonHoacBang = "nhỏ hơn hoặc bằng"
	KeywordThi            = "thì"
	KeywordKetThuc        = "kết thúc"
	KeywordVa             = "và"
	KeywordHoac           = "hoặc"
	KeywordBang           = "bằng"
	KeywordKhac           = "khác"
	KeywordTraVe          = "trả về"
)

var Keywords = map[string]string{
	"hàm":  KeywordHam,
	"biến": KeywordBien,
	"nếu":  KeywordNeu,
	"và":   KeywordVa,
	"thì":  KeywordThi,
	"hoặc": KeywordKhac,
	"bằng": KeywordBang,
	"khác": KeywordKhac,
	// Multi-word keywords are handled in the lexer
}
