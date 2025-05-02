package main

const (
	PrimitiveAny  = "tuỳ"
	PrimitiveVoid = "rỗng"
	PrimitiveB1   = "B1"
	PrimitiveN32  = "N32"
	PrimitiveN64  = "N64"
	PrimitiveZ32  = "Z32"
	PrimitiveZ64  = "Z64"
	PrimitiveR32  = "R32"
	PrimitiveR64  = "R64"
	PrimitiveC8   = "C8"
	PrimitiveC16  = "C16"
	PrimitiveC32  = "C32"
	PrimitiveS8   = "S8"
	PrimitiveS16  = "S16"
	PrimitiveS32  = "S32"
)

const (
	ContainerArray   = "mảng"
	ContainerMatrix  = "ma_trận"
	ContainerHashMap = "bảng_băm"
)

var Primitives = map[string]string{
	"B1":   PrimitiveB1,  // boolean
	"N32":  PrimitiveN32, // unsigned int
	"N64":  PrimitiveN64, // unsigned long long
	"Z32":  PrimitiveZ32, // signed int
	"Z64":  PrimitiveZ64, // signed long long
	"R32":  PrimitiveR32, // float 32 bit
	"R64":  PrimitiveR64, // float 64 bit (double)
	"số":   PrimitiveR64, // double on default (most people do this anyway)
	"C8":   PrimitiveC8,  // Char UTF-8
	"C16":  PrimitiveC16, // Char UTF-16
	"C32":  PrimitiveC32, // Char UTF-32
	"S8":   PrimitiveS8,  // String UTF-8
	"S16":  PrimitiveS16, // String UTF-16
	"S32":  PrimitiveS32, // String UTF-32
	"dãy":  PrimitiveS32,
	"rỗng": PrimitiveVoid,
}

var Containers = map[string]string{
	"mảng":     ContainerArray,
	"ma_trận":  ContainerMatrix,
	"bảng_băm": ContainerHashMap,
}
