package value

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type IGetter interface {
	GetValue(key string) reflect.Value
}

func Get(object reflect.Value, key string) reflect.Value {

	switch object.Kind() {
	case reflect.Ptr:
		if !object.IsNil() {
			v, ok := object.Interface().(IGetter)
			if ok {
				return v.GetValue(key)
			}
			return Get(object.Elem(), key)
		}
	case reflect.Map:
		return object.MapIndex(reflect.ValueOf(key))
	case reflect.Struct:
		if object.CanAddr() {
			v, ok := object.Addr().Interface().(IGetter)
			if ok {
				return v.GetValue(key)
			}
		}
		return object.FieldByName(key)
	case reflect.Interface:
		if !object.IsNil() {
			v, ok := object.Interface().(IGetter)
			if ok {
				return v.GetValue(key)
			}
		}
	}

	return reflect.ValueOf(nil)
}

func GetWithKeys(object reflect.Value, keys []string) reflect.Value {

	var v = object

	for _, key := range keys {
		v = Get(v, key)
	}

	return v
}

func Set(object reflect.Value, key string, value reflect.Value) {
	SetWithKeyIndex(object, []string{key}, 0, value)
}

func SetWithKeyIndex(object reflect.Value, keys []string, i int, value reflect.Value) {

	if i < len(keys) {

		key := keys[i]

		v := Get(object, key)

		if v.IsValid() {

			switch v.Kind() {
			case reflect.Map:
				if v.IsNil() {
					vv := reflect.MakeMap(v.Type())
					SetWithKeyIndex(vv, keys, i+1, value)
					v.Set(vv)
				} else {
					SetWithKeyIndex(v, keys, i+1, value)
				}
			case reflect.Interface:
				if v.IsNil() {
					vv := reflect.ValueOf(map[string]interface{}{})
					SetWithKeyIndex(vv, keys, i+1, value)
					v.Set(vv)
				} else {
					SetWithKeyIndex(v, keys, i+1, value)
				}
			case reflect.Slice:

				if v.IsNil() {
					v.Set(reflect.MakeSlice(v.Type(), 0, 0))
				}

				switch v.Type().Elem().Kind() {
				case reflect.Ptr:
					vv := reflect.New(v.Type().Elem().Elem())
					SetWithKeyIndex(vv, keys, i+1, value)
					v.Set(reflect.Append(v, vv))
				case reflect.Interface:
					var vvv interface{} = nil
					vv := reflect.ValueOf(vvv)
					SetWithKeyIndex(vv, keys, i+1, value)
					v.Set(reflect.Append(v, vv))
				default:
					vv := reflect.New(v.Type().Elem())
					SetWithKeyIndex(vv, keys, i+1, value)
					v.Set(reflect.Append(v, vv.Elem()))
				}
			case reflect.Ptr:
				if v.IsNil() {
					vv := reflect.New(v.Type().Elem())
					SetWithKeyIndex(vv, keys, i+1, value)
					v.Set(vv)
				} else {
					SetWithKeyIndex(v, keys, i+1, value)
				}
			default:
				SetWithKeyIndex(v, keys, i+1, value)
			}

		} else {

			switch object.Kind() {
			case reflect.Map:
				switch object.Type().Elem().Kind() {
				case reflect.Ptr:
					v = reflect.New(object.Type().Elem().Elem())
					SetWithKeyIndex(v, keys, i+1, value)
					object.SetMapIndex(reflect.ValueOf(key), v)
				case reflect.Struct:
					v = reflect.New(object.Type().Elem()).Elem()
					SetWithKeyIndex(v, keys, i+1, value)
					object.SetMapIndex(reflect.ValueOf(key), v)
				case reflect.String:
					object.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(StringValue(value, "")))
				case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
					object.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(IntValue(value, 0)))
				case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
					object.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(UintValue(value, 0)))
				case reflect.Float32, reflect.Float64:
					object.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(FloatValue(value, 0)))
				case reflect.Interface:
					object.SetMapIndex(reflect.ValueOf(key), value)
				}
			}

		}

	} else {
		SetValue(object, value)
	}

}

func SetWithKeys(object reflect.Value, keys []string, value reflect.Value) {
	SetWithKeyIndex(object, keys, 0, value)
}

func IntValue(value reflect.Value, defaultValue int64) int64 {

	var v = value

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(v.Uint())
	case reflect.Float32, reflect.Float64:
		return int64(v.Float())
	case reflect.Bool:
		if v.Bool() {
			return 1
		} else {
			return 0
		}
	case reflect.String:
		{
			vv := v.String()

			if strings.HasPrefix(vv, "0x") {
				var vvv, _ = strconv.ParseInt(vv[2:], 16, 64)
				return vvv
			} else if strings.HasPrefix(vv, "0") {
				var vvv, _ = strconv.ParseInt(vv[1:], 8, 64)
				return vvv
			} else {
				var vvv, _ = strconv.ParseInt(vv, 10, 64)
				return vvv
			}
		}
	}

	return defaultValue
}

func UintValue(value reflect.Value, defaultValue uint64) uint64 {

	var v = value

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return uint64(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint()
	case reflect.Float32, reflect.Float64:
		return uint64(v.Float())
	case reflect.Bool:
		if v.Bool() {
			return 1
		} else {
			return 0
		}
	case reflect.String:
		{
			vv := v.String()
			if strings.HasPrefix(vv, "0x") {
				var vvv, _ = strconv.ParseUint(vv[2:], 16, 64)
				return vvv
			} else if strings.HasPrefix(vv, "0") {
				var vvv, _ = strconv.ParseUint(vv[1:], 8, 64)
				return vvv
			} else {
				var vvv, _ = strconv.ParseUint(vv, 10, 64)
				return vvv
			}
		}
	}

	return defaultValue

}

func FloatValue(value reflect.Value, defaultValue float64) float64 {

	var v = value

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint())
	case reflect.Float32, reflect.Float64:
		return v.Float()
	case reflect.Bool:
		if v.Bool() {
			return 1
		} else {
			return 0
		}
	case reflect.String:
		{
			var vv, _ = strconv.ParseFloat(v.String(), 64)
			return vv
		}
	}

	return defaultValue
}

func BooleanValue(value reflect.Value, defaultValue bool) bool {

	var v = value

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() == 0 {
			return false
		} else {
			return true
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v.Uint() == 0 {
			return false
		} else {
			return true
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() == 0 {
			return false
		} else {
			return true
		}
	case reflect.Bool:
		return v.Bool()
	case reflect.String:
		{
			var vv = v.String()
			return vv == "true" || vv == "yes" || vv == "1"
		}
	}

	return defaultValue
}

func StringValue(value reflect.Value, defaultValue string) string {

	var v = value

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%u", v.Uint())
	case reflect.Float32, reflect.Float64:
		{
			vv := v.Float()
			if float64(int64(vv)) == vv {
				return fmt.Sprintf("%.0f", vv)
			} else {
				return fmt.Sprintf("%f", vv)
			}
		}
	case reflect.Bool:
		if v.Bool() {
			return "true"
		} else {
			return "false"
		}
	case reflect.String:
		return v.String()
	case reflect.Interface:
		if !v.IsNil() {
			vv, ok := v.Interface().(string)
			if ok {
				return vv
			}
		}
	}

	return defaultValue
}

func SetValue(object reflect.Value, value reflect.Value) {

	var v = object

	if v.Kind() == reflect.Func {

		fn, ok := v.Interface().(func(value reflect.Value))

		if ok && fn != nil {
			fn(value)
		}

		return
	}

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if !v.CanSet() {
		return
	}

	switch v.Kind() {
	case value.Kind():
		v.Set(value)
	case reflect.Interface:
		v.Set(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(IntValue(value, 0))
	case reflect.Float32, reflect.Float64:
		v.SetFloat(FloatValue(value, 0))
	case reflect.Bool:
		v.SetBool(BooleanValue(value, false))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(UintValue(value, 0))
	case reflect.String:
		v.SetString(StringValue(value, ""))
	case reflect.Slice:

		if v.IsNil() {
			v.Set(reflect.MakeSlice(v.Type(), 0, 0))
		}

		switch v.Type().Elem().Kind() {
		case value.Kind(), reflect.Interface:
			v.Set(reflect.Append(v, value))
		case reflect.String:
			v.Set(reflect.Append(v, reflect.ValueOf(StringValue(value, ""))))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v.Set(reflect.Append(v, reflect.ValueOf(IntValue(value, 0))))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			v.Set(reflect.Append(v, reflect.ValueOf(UintValue(value, 0))))
		case reflect.Float32, reflect.Float64:
			v.Set(reflect.Append(v, reflect.ValueOf(FloatValue(value, 0))))
		}
	}
}

func EachObject(object reflect.Value, fn func(key reflect.Value, value reflect.Value) bool) {

	var v = object

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		for _, key := range v.MapKeys() {
			if fn(key, v.MapIndex(key)) == false {
				break
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if fn(reflect.ValueOf(i), v.Index(i)) == false {
				break
			}
		}
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			if fn(reflect.ValueOf(t.Field(i).Name), v.Field(i)) == false {
				break
			}
		}
	}
}
