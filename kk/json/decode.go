package json

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func decodeNil(value reflect.Value) error {
	var v = value
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Interface:
		v.Set(reflect.ValueOf(nil))
	case reflect.Int:
	case reflect.Int8:
	case reflect.Int16:
	case reflect.Int32:
	case reflect.Int64:
		v.SetInt(0)
	case reflect.Float32:
	case reflect.Float64:
		v.SetFloat(0)
	case reflect.Array:
		v.Set(reflect.ValueOf(nil))
	case reflect.Map:
		v.Set(reflect.ValueOf(nil))
	case reflect.Bool:
		v.SetBool(false)
	case reflect.Uint:
	case reflect.Uint8:
	case reflect.Uint16:
	case reflect.Uint32:
	case reflect.Uint64:
		v.SetUint(0)
	case reflect.String:
		v.SetString("")
	}
	return nil
}

func decodeString(vv string, value reflect.Value) error {
	var v = value

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Interface:
		v.Set(reflect.ValueOf(vv))
	case reflect.Int:
	case reflect.Int8:
	case reflect.Int16:
	case reflect.Int32:
	case reflect.Int64:
		{
			if strings.HasPrefix(vv, "0x") {
				var vvv, _ = strconv.ParseInt(vv[2:], 16, 64)
				v.SetInt(vvv)
			} else if strings.HasPrefix(vv, "0") {
				var vvv, _ = strconv.ParseInt(vv[1:], 8, 64)
				v.SetInt(vvv)
			} else {
				var vvv, _ = strconv.ParseInt(vv, 10, 64)
				v.SetInt(vvv)
			}

		}
	case reflect.Float32:
	case reflect.Float64:
		{
			var vvv, _ = strconv.ParseFloat(vv, 64)
			v.SetFloat(vvv)
		}
	case reflect.Array:
		v.Set(reflect.ValueOf(nil))
	case reflect.Map:
		v.Set(reflect.ValueOf(nil))
	case reflect.Bool:
		if vv == "true" || vv == "yes" || vv == "1" {
			v.SetBool(true)
		} else {
			v.SetBool(false)
		}
	case reflect.Uint:
	case reflect.Uint8:
	case reflect.Uint16:
	case reflect.Uint32:
	case reflect.Uint64:
		{
			var vvv, _ = strconv.ParseUint(vv, 10, 64)
			v.SetUint(vvv)
		}
	case reflect.String:
		v.SetString(vv)
	}
	return nil
}

func decodeFloat64(vv float64, value reflect.Value) error {
	var v = value
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Interface:
		v.Set(reflect.ValueOf(vv))
	case reflect.Int:
	case reflect.Int8:
	case reflect.Int16:
	case reflect.Int32:
	case reflect.Int64:
		v.SetInt(int64(vv))
	case reflect.Float32:
	case reflect.Float64:
		v.SetFloat(vv)
	case reflect.Array:
		v.Set(reflect.ValueOf(nil))
	case reflect.Map:
		v.Set(reflect.ValueOf(nil))
	case reflect.Bool:
		if vv != 0 {
			v.SetBool(true)
		} else {
			v.SetBool(false)
		}
	case reflect.Uint:
	case reflect.Uint8:
	case reflect.Uint16:
	case reflect.Uint32:
	case reflect.Uint64:
		v.SetUint(uint64(vv))
	case reflect.String:
		if float64(int64(vv)) == vv {
			v.SetString(fmt.Sprintf("%.0f", vv))
		} else {
			v.SetString(fmt.Sprintf("%f", vv))
		}
	}

	return nil
}

func decodeBoolean(vv bool, value reflect.Value) error {
	var v = value
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Interface:
		v.Set(reflect.ValueOf(vv))
	case reflect.Int:
	case reflect.Int8:
	case reflect.Int16:
	case reflect.Int32:
	case reflect.Int64:
		if vv {
			v.SetInt(1)
		} else {
			v.SetInt(0)
		}
	case reflect.Float32:
	case reflect.Float64:
		if vv {
			v.SetFloat(1)
		} else {
			v.SetFloat(0)
		}
	case reflect.Array:
		v.Set(reflect.ValueOf(nil))
	case reflect.Map:
		v.Set(reflect.ValueOf(nil))
	case reflect.Bool:
		v.SetBool(vv)
	case reflect.Uint:
	case reflect.Uint8:
	case reflect.Uint16:
	case reflect.Uint32:
	case reflect.Uint64:
		if vv {
			v.SetUint(1)
		} else {
			v.SetUint(0)
		}
	case reflect.String:
		if vv {
			v.SetString("true")
		} else {
			v.SetString("false")
		}
	}
	return nil
}

func decodeObject(dec *json.Decoder, value reflect.Value) error {

	var err error = nil
	var v = value

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var mapvalue map[string]interface{} = nil
	var fdvalue map[string]reflect.Value = nil

	switch v.Kind() {
	case reflect.Interface:
	case reflect.Map:
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

			case json.Delim:
				switch token.(json.Delim).String() {
				case "}":
					return nil
				default:
					return errors.New("3 json decodeObject type: " + value.String())
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

	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	var vv []interface{} = nil
	var valueType reflect.Type = nil

	switch v.Kind() {
	case reflect.Array:
		valueType = v.Type().Elem()
		vv = reflect.New(v.Type()).Interface().([]interface{})
	case reflect.Interface:
		vv = []interface{}{}
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

			switch token.(type) {
			case json.Delim:
				switch token.(json.Delim).String() {
				case "]":
					return nil
				}
			}

			var vvv interface{} = nil

			if valueType != nil {
				vvv = reflect.New(valueType)
			}

			err = decodeToken(dec, token, reflect.ValueOf(&vvv))

			if err != nil {
				return err
			}

			if vv != nil {
				vv = append(vv, vvv)
			}
		}

		return nil
	}

	err = fn()

	if err != nil {
		return err
	}

	if vv != nil {
		v.Set(reflect.ValueOf(vv))
	}

	return nil
}

func decodeToken(dec *json.Decoder, token json.Token, value reflect.Value) error {

	if token == nil {
		return decodeNil(value)
	}

	switch token.(type) {
	case string:
		return decodeString(token.(string), value)
	case float64:
		return decodeFloat64(token.(float64), value)
	case bool:
		return decodeBoolean(token.(bool), value)
	case json.Delim:
		var d = token.(json.Delim)
		switch d.String() {
		case "{":
			return decodeObject(dec, value)
		case "[":
			return decodeArray(dec, value)
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
