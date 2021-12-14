package revip

import (
	"fmt"
	"reflect"
)

func Postprocess(c Config, op ...Option) error {
	return postprocess(c, nil, op)
}

func postprocessApply(c Config, path []string, op []Option) error {
	var err error
	for _, f := range op {
		err = f(c, path)
		if err != nil {
			return err
		}
	}
	return nil
}

func postprocess(c Config, path []string, op []Option) error {
	err := postprocessApply(c, path, op)
	if err != nil {
		return err
	}

	//

	kind := reflect.TypeOf(c).Kind()
	value := reflect.ValueOf(c)

	//

	switch kind {
	case reflect.Ptr:
		if value.IsNil() {
			return nil // NOTE: skip nil's, this mean we don't have a default value
		}

		// FIXME: will call op twice if receiver is not pointer, not sure how to fix atm
		v := indirectValue(value)
		return postprocess(
			v.Interface(),
			path,
			op,
		)
	case reflect.Struct:
		return walkStruct(c, func(v reflect.Value, xs []string) error {
			return postprocess(
				v.Interface(),
				append(path, xs...),
				op,
			)
		})
	case reflect.Array, reflect.Slice:
		for n := 0; n < value.Len(); n++ {
			err := postprocess(
				value.Index(n).Interface(),
				append(path, fmt.Sprintf("[%d]", n)),
				op,
			)
			if err != nil {
				return err
			}
		}
	case reflect.Map:
		for _, k := range value.MapKeys() {
			err := postprocess(
				value.MapIndex(k).Interface(),
				append(path, fmt.Sprintf("[%q]", k.String())),
				op,
			)
			if err != nil {
				return err
			}

		}
	default:
		return nil
	}

	return nil
}

//

func WithDefaults() Option {
	return func(c Config, m ...OptionMeta) error {
		var err error

		v, ok := c.(Defaultable)
		if ok && !isnil(reflect.ValueOf(v)) {
			err = expectKind(reflect.TypeOf(v), reflect.Ptr)
			if err != nil {
				return err
			}

			v.Default()
		}
		return nil
	}
}

func WithValidation() Option {
	return func(c Config, m ...OptionMeta) error {
		var err error

		v, ok := c.(Validatable)
		if ok && !isnil(reflect.ValueOf(v)) {
			err = expectKind(reflect.TypeOf(v), reflect.Ptr)
			if err != nil {
				return err
			}

			err = v.Validate()
			if err != nil {
				var path []string
				if len(m) > 0 {
					path = m[0].([]string)
				}
				return &ErrPostprocess{
					Type: reflect.TypeOf(c).String(),
					Path: path,
					Err:  err,
				}
			}
		}
		return nil
	}
}

func WithExpansion() Option {
	return func(c Config, m ...OptionMeta) error {
		var err error

		v, ok := c.(Expandable)
		if ok && !isnil(reflect.ValueOf(v)) {
			err = expectKind(reflect.TypeOf(v), reflect.Ptr)
			if err != nil {
				return err
			}

			err = v.Expand()
			if err != nil {
				var path []string
				if len(m) > 0 {
					path = m[0].([]string)
				}
				return &ErrPostprocess{
					Type: reflect.TypeOf(c).String(),
					Path: path,
					Err:  err,
				}
			}
		}
		return nil
	}
}
