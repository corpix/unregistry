package server

import (
	"encoding/json"
	"fmt"
	"strings"

	"git.backbone/corpix/unregistry/pkg/errors"
	"git.backbone/corpix/unregistry/pkg/reflect"
)

const (
	EncodeHeadersTag              = "header"
	EncodeHeadersTagOptsDelimiter = ","

	EncodeHeadersTagFormatJSON = "json"
)

func EncodeHeaders(h Headers, v interface{}) error {
	var (
		rt = reflect.IndirectType(reflect.TypeOf(v))
		rv = reflect.IndirectValue(reflect.ValueOf(v))
		l  = rt.NumField()

		ft        reflect.StructField
		key       string
		keyOpts   []string
		format    string
		formatErr error

		vf    reflect.Value
		fv    interface{}
		value string
	)

loop:
	for n := 0; n < l; n++ {
		ft = rt.Field(n)
		key = ft.Tag.Get(EncodeHeadersTag)
		keyOpts = strings.Split(key, EncodeHeadersTagOptsDelimiter)
		format = ""

		//

		if len(keyOpts) > 1 {
			key = keyOpts[0]
			format = keyOpts[1]
		}
		if key == "" || key == "-" {
			continue
		}

		//

		vf = rv.Field(n)
		fv = vf.Interface()

		//

		switch format {
		case EncodeHeadersTagFormatJSON:
			if reflect.IsNil(vf) {
				continue loop
			}

			buf, err := json.Marshal(fv)
			if err != nil {
				formatErr = err
				goto fail
			}
			value = string(buf)
		default:
			switch v := fv.(type) {
			case []string:
				value = strings.Join(v, ",")
			default:
				value = fmt.Sprintf("%s", v)
			}
		}

		//

		if value == "" {
			continue loop
		}
		h.Add(key, value)

		//

	fail:
		if formatErr != nil {
			return errors.Wrapf(
				formatErr, "failed to marshal format %q field %q into header %q with value %#v",
				format, ft.Name, key, fv,
			)
		}
	}

	return nil
}
