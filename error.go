package main

import "fmt"

type ErrorID int

const (
	ExpectToken ErrorID = iota
	UnexpectedToken
	WrongToken
	TypeMismatch
	UndeclaredIdentifier
	MissingReturn
	InvalidFunctionCall
)

var errorMessagesVi = map[ErrorID]string{
	ExpectToken:          "Mong đợi %v ở vị trí này",
	UnexpectedToken:      "Không mong đợi ký hiệu '%v' ở vị trí này.",
	WrongToken:           "Mong đợi ký hiệu '%v' thay vì '%v' ở vị trí này.",
	TypeMismatch:         "Không thể gán kiểu '%v' cho '%v'.",
	UndeclaredIdentifier: "Không tìm thấy định danh '%v'.",
	MissingReturn:        "Thiếu câu lệnh trả về trong hàm có kiểu trả về '%v'.",
	InvalidFunctionCall:  "Hàm '%v' không tồn tại hoặc sai kiểu tham số.",
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
