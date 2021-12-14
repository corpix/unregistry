package reflect

import (
	"reflect"
)

type (
	Type        = reflect.Type
	Kind        = reflect.Kind
	Value       = reflect.Value
	StructField = reflect.StructField
)

const (
	Invalid       = reflect.Invalid
	Bool          = reflect.Bool
	Int           = reflect.Int
	Int8          = reflect.Int8
	Int16         = reflect.Int16
	Int32         = reflect.Int32
	Int64         = reflect.Int64
	Uint          = reflect.Uint
	Uint8         = reflect.Uint8
	Uint16        = reflect.Uint16
	Uint32        = reflect.Uint32
	Uint64        = reflect.Uint64
	Uintptr       = reflect.Uintptr
	Float32       = reflect.Float32
	Float64       = reflect.Float64
	Complex64     = reflect.Complex64
	Complex128    = reflect.Complex128
	Array         = reflect.Array
	Chan          = reflect.Chan
	Func          = reflect.Func
	Interface     = reflect.Interface
	Map           = reflect.Map
	Ptr           = reflect.Ptr
	Slice         = reflect.Slice
	String        = reflect.String
	Struct        = reflect.Struct
	UnsafePointer = reflect.UnsafePointer
)

var (
	TypeOf  = reflect.TypeOf
	ValueOf = reflect.ValueOf
)

func CheckValue(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return NewErrPtrRequired(v)
	}

	return nil
}

func CheckStruct(reflectValue reflect.Value) error {
	if reflectValue.Kind() != reflect.Struct {
		return NewErrUnexpecterKind(
			reflectValue.Kind(),
			reflect.Struct,
		)
	}

	return nil
}

func StructFieldExported(field reflect.StructField) bool {
	// FIXME: thats a terrible way, to check if field is exported, could we do better?
	// From reflect docs:
	// PkgPath is the package path that qualifies a lower case (unexported)
	// field name. It is empty for upper case (exported) field names.
	// See https://golang.org/ref/spec#Uniqueness_of_identifiers
	return field.PkgPath == ""
}

func IndirectValue(reflectValue reflect.Value) reflect.Value {
	if reflectValue.Kind() == reflect.Ptr {
		return reflectValue.Elem()
	}
	return reflectValue
}

func IndirectType(reflectType reflect.Type) reflect.Type {
	if reflectType.Kind() == reflect.Ptr || reflectType.Kind() == reflect.Slice {
		return reflectType.Elem()
	}
	return reflectType
}

func ExpectKind(reflectType reflect.Type, ks ...reflect.Kind) error {
	k := reflectType.Kind()

	for _, ek := range ks {
		if ek == k {
			return nil
		}
	}

	return &ErrUnexpectedKind{
		Got:      k,
		Expected: ks,
	}
}

func IsNil(reflectValue reflect.Value) bool {
	switch reflectValue.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice:
		return reflectValue.IsNil()
	default:
		return false
	}

}
