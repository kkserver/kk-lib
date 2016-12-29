package logic

import (
	"github.com/kkserver/kk-lib/kk/app"
	Value "github.com/kkserver/kk-lib/kk/value"
	"reflect"
	"strings"
)

const ERROR_UNKNOWN = 0xff00

var ErrnoKeys = []string{"output", "errno"}
var ErrmsgKeys = []string{"output", "errmsg"}
var ResultKeys = []string{"result"}
var ObjectKeys = []string{"object"}
var InputKeys = []string{"input"}
var OutputKeys = []string{"output"}
var ViewKeys = []string{"view"}

type ILogic interface {
	Exec(a app.IApp, program IProgram, ctx IContext) error
}

type IProgram interface {
	GetLogic(name string) ILogic
}

type Program struct {
}

type IContext interface {
	Begin()
	End()
	Set(keys []string, value interface{})
	Get(keys []string) interface{}
	ReflectValue(value interface{}) interface{}
}

type Context struct {
	values []map[string]interface{}
}

func (P *Program) GetLogic(name string) ILogic {
	return nil
}

func Exec(a app.IApp, program IProgram, ctx IContext) error {

	v := program.GetLogic("In")

	if v == nil {
		return app.NewError(ERROR_UNKNOWN, "Not Found In Logic")
	}

	return v.Exec(a, program, ctx)
}

func (C *Context) Begin() {
	if C.values == nil {
		C.values = []map[string]interface{}{map[string]interface{}{}}
	} else {
		C.values = append(C.values, map[string]interface{}{})
	}
}

func (C *Context) End() {
	if C.values != nil && len(C.values) > 1 {
		C.values = C.values[0 : len(C.values)-1]
	}
}

func (C *Context) Set(keys []string, value interface{}) {
	if C.values != nil && len(C.values) > 0 {
		vs := C.values[len(C.values)-1]
		Value.SetWithKeys(reflect.ValueOf(vs), keys, reflect.ValueOf(value))
	}

}

func (C *Context) Get(keys []string) interface{} {
	if C.values != nil && len(C.values) > 0 {
		idx := len(C.values) - 1
		for idx >= 0 {
			vs := C.values[idx]
			v := Value.GetWithKeys(reflect.ValueOf(vs), keys)
			if v.IsValid() && v.CanInterface() && !v.IsNil() {
				return v.Interface()
			}
			idx = idx - 1
		}
	}
	return nil
}

func (C *Context) ReflectValue(value interface{}) interface{} {

	v := reflect.ValueOf(value)

	if v.Kind() == reflect.String {
		s := v.String()
		if strings.HasPrefix(s, "=") {
			return C.Get(strings.Split(s[1:], "."))
		}
	}

	if v.CanInterface() {
		return v.Interface()
	}

	return nil
}
