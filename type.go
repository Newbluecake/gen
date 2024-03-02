package gen

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/go-clang/bootstrap/clang"
)

// Defines all available Go types.
const (
	GoByte      = "byte"
	GoInt8      = "int8"
	GoUInt8     = "uint8"
	GoInt16     = "int16"
	GoUInt16    = "uint16"
	GoInt32     = "int32"
	GoUInt32    = "uint32"
	GoInt64     = "int64"
	GoUInt64    = "uint64"
	GoFloat32   = "float32"
	GoFloat64   = "float64"
	GoBool      = "bool"
	GoInterface = "interface"
	GoPointer   = "unsafe.Pointer"
)

// Defines all available C types.
const (
	CChar      = "char"
	CSChar     = "schar"
	CUChar     = "uchar"
	CShort     = "short"
	CUShort    = "ushort"
	CInt       = "int"
	CUInt      = "uint"
	CLongInt   = "long"
	CULongInt  = "ulong"
	CLongLong  = "longlong"
	CULongLong = "ulonglong"
	CFloat     = "float"
	CDouble    = "double"
)

// Type represents a generation type.
type Type struct {
	// CName C Type name
	CName string

	// CGoName Cgo Type name
	CGoName string

	// GoName Go Type name
	GoName string

	// LengthOfSlice length of slice
	LengthOfSlice string

	// ArraySize size of array
	ArraySize int64

	// PointerLevel level of pointer
	PointerLevel int

	// IsPrimitive whether the this Type is primitive
	IsPrimitive bool

	// IsArray whether the this Type is array
	IsArray bool

	// IsEnumLiteral whether the this Type is enum literal
	IsEnumLiteral bool

	// IsFunctionPointer whether the this Type is function pointer
	IsFunctionPointer bool

	// IsReturnArgument whether the this Type is return argument
	IsReturnArgument bool

	// IsSlice whether the this Type is slice
	IsSlice bool

	// IsPointerComposition whether the this Type is pointer composition
	IsPointerComposition bool

	// IsBitField whether the this Type is bit field
	IsBitField bool
}

// TypeFromClangType returns the Type from Clang type.
func TypeFromClangType(cType clang.Type) (Type, error) {
	typ := Type{
		CName:             cType.Spelling(),
		PointerLevel:      0,
		IsPrimitive:       true,
		IsArray:           false,
		ArraySize:         -1,
		IsEnumLiteral:     false,
		IsFunctionPointer: false,
		IsBitField:        false,
	}

	switch cType.Kind() {
	case clang.Type_Char_S:
		typ.CGoName = CSChar
		typ.GoName = GoInt8

	// FIXME: I guess UChar and Char_U are the same, but I'm not sure
	case clang.Type_Char_U, clang.Type_UChar:
		typ.CGoName = CUChar
		typ.GoName = GoUInt8

	case clang.Type_Int:
		typ.CGoName = CInt
		typ.GoName = GoInt32

	case clang.Type_Short:
		typ.CGoName = CShort
		typ.GoName = GoInt16

	case clang.Type_UShort:
		typ.CGoName = CUShort
		typ.GoName = GoUInt16

	case clang.Type_UInt:
		typ.CGoName = CUInt
		typ.GoName = GoUInt32

	case clang.Type_Long:
		typ.CGoName = CLongInt
		typ.GoName = GoInt64

	case clang.Type_ULong:
		typ.CGoName = CULongInt
		typ.GoName = GoUInt64

	case clang.Type_LongLong:
		typ.CGoName = CLongLong
		typ.GoName = GoInt64

	case clang.Type_ULongLong:
		typ.CGoName = CULongLong
		typ.GoName = GoUInt64

	case clang.Type_Float:
		typ.CGoName = CFloat
		typ.GoName = GoFloat32

	case clang.Type_Double:
		typ.CGoName = CDouble
		typ.GoName = GoFloat64

	case clang.Type_Bool:
		typ.GoName = GoBool

	case clang.Type_Void:
		// TODO(go-clang): does not exist in Go, what should we do with it? https://github.com/go-clang/gen/issues/50
		typ.CGoName = "void"
		typ.GoName = "void"

	case clang.Type_ConstantArray:
		subTyp, err := TypeFromClangType(cType.ArrayElementType())
		if err != nil {
			return Type{}, err
		}

		typ.CGoName = subTyp.CGoName
		typ.GoName = subTyp.GoName
		typ.PointerLevel += subTyp.PointerLevel
		typ.IsArray = true
		typ.ArraySize = cType.ArraySize()

	case clang.Type_Typedef:
		typ.IsPrimitive = false

		typeStr := cType.Spelling()
		switch typeStr {
		case "CXString": // TODO(go-clang): eliminate CXString from the generic code https://github.com/go-clang/gen/issues/25
			typeStr = "cxstring"

		case "time_t":
			typ.CGoName = typeStr
			typeStr = "time.Time"
			typ.IsPrimitive = true

		default:
			typeStr = TrimLanguagePrefix(cType.Declaration().Type().Spelling())
		}

		typ.CGoName = cType.Declaration().Type().Spelling()
		typ.GoName = typeStr

		if cType.CanonicalType().Kind() == clang.Type_Enum {
			typ.IsEnumLiteral = true
			typ.IsPrimitive = true
		}

	case clang.Type_Pointer:
		typ.PointerLevel++

		if cType.PointeeType().CanonicalType().Kind() == clang.Type_FunctionProto {
			typ.IsFunctionPointer = true
		}

		subTyp, err := TypeFromClangType(cType.PointeeType())
		if err != nil {
			return Type{}, err
		}

		typ.CGoName = subTyp.CGoName
		typ.GoName = subTyp.GoName
		typ.PointerLevel += subTyp.PointerLevel
		typ.IsPrimitive = subTyp.IsPrimitive

	case clang.Type_Record:
		typ.CGoName = cType.Declaration().Type().Spelling()
		typ.GoName = TrimLanguagePrefix(typ.CGoName)
		typ.IsPrimitive = false

	case clang.Type_FunctionProto:
		typ.IsFunctionPointer = true
		typ.CGoName = cType.Declaration().Type().Spelling()
		typ.GoName = TrimLanguagePrefix(typ.CGoName)

	case clang.Type_Enum:
		typ.GoName = TrimLanguagePrefix(cType.Declaration().DisplayName())
		typ.IsEnumLiteral = true
		typ.IsPrimitive = true

	case clang.Type_Elaborated:
		return TypeFromClangType(cType.CanonicalType())

	case clang.Type_Unexposed: // there is a bug in clang for enums the kind is set to unexposed dunno why, bug persisted since 2013: https://llvm.org/bugs/show_bug.cgi?id=15089
		subTyp, err := TypeFromClangType(cType.CanonicalType())
		if err != nil {
			return Type{}, err
		}

		typ.CGoName = subTyp.CGoName
		typ.GoName = subTyp.GoName
		typ.PointerLevel += subTyp.PointerLevel
		typ.IsPrimitive = subTyp.IsPrimitive

	default:
		return Type{}, fmt.Errorf("unhandled type %q of kind %q", cType.Spelling(), cType.Kind().Spelling())
	}

	return typ, nil
}

// ArrayNameFromLength returns the array name from lengthCName length naming.
func ArrayNameFromLength(lengthCName string) string {
	switch {
	case strings.HasPrefix(lengthCName, "num_"):
		return strings.TrimPrefix(lengthCName, "num_")

	case strings.HasPrefix(lengthCName, "num"):
		return strings.TrimPrefix(lengthCName, "num")

	case strings.HasPrefix(lengthCName, "_size"):
		return strings.TrimSuffix(lengthCName, "_size")

	default:
		if strings.HasPrefix(lengthCName, "Num") {
			pan := strings.TrimPrefix(lengthCName, "Num")
			if unicode.IsUpper(rune(pan[0])) {
				return pan
			}
		}
	}

	return ""
}

// IsInteger reports whether the typ is Go integer type.
func IsInteger(typ *Type) bool {
	switch typ.GoName {
	case GoInt8, GoUInt8, GoInt16, GoUInt16, GoInt32, GoUInt32, GoInt64, GoUInt64:
		return true
	}

	return false
}
