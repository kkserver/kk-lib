package json

import (
	"bytes"
	"encoding/json"
	"github.com/kkserver/kk-lib/kk/dynamic"
	"reflect"
)

func encode(object interface{}, w *bytes.Buffer) error {

	if object == nil {
		w.WriteString("null")
		return nil
	}

	v := reflect.ValueOf(object)

	switch v.Kind() {
	case reflect.Map:
		i := 0
		w.WriteString("{")
		for _, key := range v.MapKeys() {
			vv := v.MapIndex(key)
			if key.CanInterface() && vv.CanInterface() {
				if i != 0 {
					w.WriteString(",")
				}
				encode(dynamic.StringValue(key.Interface(), ""), w)
				w.WriteString(":")
				encode(vv.Interface(), w)
				i = i + 1
			}
		}
		w.WriteString("}")
	default:
		b, err := json.Marshal(object)
		if err != nil {
			return err
		}
		w.Write(b)
	}

	return nil
}

func Encode(object interface{}) ([]byte, error) {

	w := bytes.NewBuffer(nil)

	err := encode(object, w)

	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}
