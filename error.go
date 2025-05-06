package main

import "fmt"

type ErrorID int

const (
	ExpectToken ErrorID = iota
	UnexpectedToken
	WrongToken
	TypeMismatch
	ReturnTypeMismatch
	UndeclaredIdentifier
	MissingReturn
	InvalidFunctionCall
	ArgumentCountMismatch
	ArgumentTypeMismatch
	RedeclarationVar
	RedeclarationFunction
	UnknownExpression
	InvalidIdentifierUsage
	UnknownIdentifierType
	InvalidCasting
	ErrorBinaryExpr
	InvalidArrayAccessIndex
	InvalidArrayAccessType
	InvalidArrayAccessDim
	InvalidArrayAccessRange
)

var errorMessagesVi = map[ErrorID]string{
	ExpectToken:             "Mong đợi %v ở vị trí này.",
	UnexpectedToken:         "Không mong đợi ký hiệu '%v' ở vị trí này.",
	WrongToken:              "Mong đợi ký hiệu '%v' thay vì '%v' ở vị trí này.",
	TypeMismatch:            "Sai kiểu '%v' thay vì '%v'.",
	ReturnTypeMismatch:      "Không thể trả về giá trị kiểu '%v', mong đợi kiểu '%v'.",
	UndeclaredIdentifier:    "Không tìm thấy định danh '%v'.",
	MissingReturn:           "Thiếu câu lệnh trả về trong hàm có kiểu trả về '%v'.",
	InvalidFunctionCall:     "Hàm '%v' không tồn tại",
	ArgumentCountMismatch:   "Số lượng đối số (%v) của lời gọi hàm không khớp với số lượng tham số (%v) của hàm '%v'.",
	ArgumentTypeMismatch:    "Đối số '%v' khác kiểu với tham số '%v'.",
	RedeclarationVar:        "Lỗi khai báo lại biến '%v'.",
	RedeclarationFunction:   "Lỗi khai báo lại hàm '%v'.",
	UnknownExpression:       "Biểu thức không xác định.",
	InvalidIdentifierUsage:  "Không thể đánh giá được cách sử dụng ký hiệu '%v'.",
	UnknownIdentifierType:   "Ký hiệu không xác định.",
	InvalidCasting:          "Không thể chuyển kiểu '%v' sang kiểu '%v'.",
	ErrorBinaryExpr:         "Không thể thực hiện phép toán giữa '%v' và '%v'",
	InvalidArrayAccessIndex: "Không thể truy cập phần tử của biến này với chỉ số khác số nguyên",
	InvalidArrayAccessType:  "Không thể truy cập phần tử của biểu thức này, làm ơn truy cập một biến thuộc kiểu mảng",
	InvalidArrayAccessDim:   "Chiều của chỉ số (%d) khác với chiều của biến (%d)",
	InvalidArrayAccessRange: "Chỉ số (%d) nằm ngoài giới hạn của mảng [%d..%d]",
}

type LangError struct {
	ID       ErrorID
	Args     []any
	Line     int
	Column   int
	Language string // "vi" or "en"
}

func NewLangError(id ErrorID, args ...any) *LangError {
	return &LangError{
		ID:       id,
		Args:     args,
		Language: "vi", // default, TODO: make configurable later
	}
}

func (e *LangError) At(line, column int) *LangError {
	e.Line = line
	e.Column = column
	return e
}

func (e *LangError) Error() string {
	template, ok := errorMessagesVi[e.ID]
	if !ok {
		template = "Lỗi không xác định."
	}
	msg := fmt.Sprintf(template, e.Args...)
	return fmt.Sprintf("[Dòng %d, Cột %d] %s", e.Line, e.Column, msg)
}
