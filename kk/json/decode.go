package json

import (
	"bytes"
	"encoding/json"
	"errors"
	V "github.com/kkserver/kk-lib/kk/value"
	"reflect"
	"strings"
)

func decodeObject(dec *json.Decoder, value reflect.Value) error {

	var err error = nil
	var v = value

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var mapvalue map[string]interface{} = nil
	var fdvalue map[string]reflect.Value = nil

	switch v.Kind() {
	case reflect.Interface, reflect.Map:
		mapvalue = map[string]interface{}{}
	case reflect.Struct:
		fdvalue = map[string]reflect.Value{}
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			fd := t.Field(i)
			tag := fd.Tag.Get("json")
			name := fd.Name
			if tag != "" {
				name = strings.Split(tag, ",")[0]
				if name == "-" {
					continue
				}
			}
			fdvalue[name] = v.Field(i)
		}
	}

	fn := func() error {

		for dec.More() {

			token, err := dec.Token()

			if err != nil {
				return err
			}

			if token == nil {
				return errors.New("1 json decodeObject type: " + value.String())
			}

			switch token.(type) {
			case string:

				key := token.(string)

				if dec.More() {

					token, err := dec.Token()

					if err != nil {
						return err
					}

					if mapvalue != nil {
						var vv interface{} = nil
						err = decodeToken(dec, token, reflect.ValueOf(&vv))
						if err != nil {
							return err
						}
						mapvalue[key] = vv
					} else if fdvalue != nil {
						fd, ok := fdvalue[key]
						if ok {

							if fd.Kind() == reflect.Ptr {
								fd.Set(reflect.New(fd.Type().Elem()))
							}

							err = decodeToken(dec, token, fd)

							if err != nil {
								return err
							}
						} else {
							var vv interface{} = nil
							err = decodeToken(dec, token, reflect.ValueOf(&vv))
							if err != nil {
								return err
							}
						}
					}

				} else {
					return errors.New("2 json decodeObject type: " + value.String())
				}
			default:
				return errors.New("4 json decodeObject type: " + value.String())
			}
		}

		return nil
	}

	err = fn()

	if err != nil {
		return err
	}

	if mapvalue != nil {
		v.Set(reflect.ValueOf(mapvalue))
	}

	return nil
}

func decodeArray(dec *json.Decoder, value reflect.Value) error {

	var err error = nil
	var v = value

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var vv reflect.Value
	var valueType reflect.Type = nil

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		valueType = v.Type().Elem()
		vv = reflect.New(v.Type()).Elem()
	case reflect.Interface:
		vv = reflect.ValueOf([]interface{}{})
		valueType = vv.Type().Elem()
	}

	fn := func() error {

		for dec.More() {

			token, err := dec.Token()

			if err != nil {
				return err
			}

			if token == nil {
				return errors.New("json decodeArray type: " + value.Type().Name())
			}

			var vvv reflect.Value

			if valueType != nil {
				vvv = reflect.New(valueType).Elem()
			}

			err = decodeToken(dec, token, vvv)

			if err != nil {
				return err
			}

			switch vv.Kind() {
			case reflect.Slice, reflect.Array:
				vv = reflect.Append(vv, vvv)
			}
		}

		return nil
	}

	err = fn()

	if err != nil {
		return err
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.Interface:
		v.Set(vv)
	}

	return nil
}

func decodeToken(dec *json.Decoder, token json.Token, value reflect.Value) error {

	if token == nil {
		V.SetValue(value, reflect.ValueOf(nil))
		return nil
	}

	switch token.(type) {
	case string:
		V.SetValue(value, reflect.ValueOf(token.(string)))
		return nil
	case float64:
		V.SetValue(value, reflect.ValueOf(token.(float64)))
		return nil
	case bool:
		V.SetValue(value, reflect.ValueOf(token.(bool)))
		return nil
	case json.Delim:
		var d = token.(json.Delim)
		switch d.String() {
		case "{":
			err := decodeObject(dec, value)
			if err != nil {
				return err
			}
			_, err = dec.Token()
			return err
		case "[":
			err := decodeArray(dec, value)
			if err != nil {
				return err
			}
			_, err = dec.Token()
			return err
		default:
			return errors.New("json decodeToken : " + d.String())
		}
	}

	return nil
}

func decode(dec *json.Decoder, value reflect.Value) error {

	if dec.More() {
		token, err := dec.Token()
		if err != nil {
			return err
		}
		return decodeToken(dec, token, value)
	}

	return nil
}

func Decode(data []byte, object interface{}) error {
	rd := bytes.NewReader(data)
	dec := json.NewDecoder(rd)
	return decode(dec, reflect.ValueOf(object))
}
