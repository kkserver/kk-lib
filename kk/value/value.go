package value

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func Get(object reflect.Value, key string) reflect.Value {

	switch object.Kind() {
	case reflect.Ptr:
		if !object.IsNil() {
			return Get(object.Elem(), key)
		}
	case reflect.Map:
		return object.MapIndex(reflect.ValueOf(key))
	case reflect.Struct:
		return object.FieldByName(key)
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

	switch object.Kind() {
	case reflect.Map:
		object.SetMapIndex(reflect.ValueOf(key), value)
	case reflect.Struct:
		SetValue(object.FieldByName(key), value)
	case reflect.Interface:
		object.Set(value)
	case reflect.Ptr:
		if object.IsNil() {
			object.Set(reflect.New(object.Type().Elem()))
		}
		Set(object.Elem(), key, value)
	}

}

func SetWithKeys(object reflect.Value, keys []string, value reflect.Value) {

	var v = object

	for i, key := range keys {

		if i+1 == len(keys) {
			Set(v, key, value)
			return
		}

		v = Get(v, key)
		switch v.Kind() {
		case reflect.Interface:
			if v.IsNil() {
				v.Set(reflect.ValueOf(map[string]interface{}{}))
			}
		case reflect.Ptr:
			if v.IsNil() {
				v.Set(reflect.New(object.Type().Elem()))
			}
		}
	}

	SetValue(v, value)

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
	}

	return defaultValue
}

func SetValue(object reflect.Value, value reflect.Value) {

	var v = object

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
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
