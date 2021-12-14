package revip

import (
	"errors"
	"reflect"
)

func indirectValue(reflectValue reflect.Value) reflect.Value {
	if reflectValue.Kind() == reflect.Ptr {
		return reflectValue.Elem()
	}
	return reflectValue
}

func indirectType(reflectType reflect.Type) reflect.Type {
	if reflectType.Kind() == reflect.Ptr || reflectType.Kind() == reflect.Slice {
		return reflectType.Elem()
	}
	return reflectType
}

func isnil(reflectValue reflect.Value) bool {
	if reflectValue.Kind() == reflect.Ptr {
		return reflectValue.IsNil()
	}
	return false
}

func expectKind(reflectType reflect.Type, ks ...reflect.Kind) error {
	k := reflectType.Kind()

	for _, ek := range ks {
		if ek == k {
			return nil
		}
	}

	return &ErrUnexpectedKind{
		Type:     reflectType,
		Got:      k,
		Expected: ks,
	}
}

//

func structFieldExported(f reflect.StructField) bool {
	return f.PkgPath == ""
}

//

var (
	stopIteration = errors.New("stop iteration")
	skipBranch    = errors.New("skip branch")
)

//

func walkStructIter(v reflect.Value, path []string, cb func(reflect.Value, []string) error) error {
	var (
		t   = v.Type()
		k   = t.Kind()
		err error
	)

	if len(path) > 0 { // do not invoke cb for root struct (it is pointless)
		err = cb(v, path)
		switch err {
		case nil:
		case skipBranch:
			return nil
		case stopIteration:
			return err
		default:
			return err
		}
	}

	switch k {
	case reflect.Ptr:
		if !v.IsNil() {
			return walkStructIter(
				indirectValue(v),
				path, cb,
			)
		}
	case reflect.Struct:
		for n := 0; n < v.NumField(); n++ {
			ff := t.Field(n)
			if !structFieldExported(ff) {
				continue
			}

			fv := v.Field(n)
			next := append(path, ff.Name)

			//

			err = walkStructIter(fv, next, cb)
			switch err {
			case nil:
			case skipBranch:
				continue
			case stopIteration:
				return err
			default:
				return err
			}
		}

		return nil
	}

	return nil
}

func walkStruct(value interface{}, cb func(reflect.Value, []string) error) error {
	v := indirectValue(reflect.ValueOf(value))

	err := expectKind(v.Type(), reflect.Struct)
	if err != nil {
		return err
	}

	err = walkStructIter(v, []string{}, cb)
	switch err {
	case nil:
	case stopIteration:
	default:
		return err
	}

	return nil
}

func prefixPath(namespace string, path []string) []string {
	return append([]string{namespace}, path...)
}
