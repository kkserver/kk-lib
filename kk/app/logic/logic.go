package logic

import (
	"errors"
	"github.com/kkserver/kk-lib/kk/app"
	Value "github.com/kkserver/kk-lib/kk/value"
	"reflect"
	"strings"
)

const ERROR_UNKNOWN = 0xff00

var ErrnoKeys = []string{"errno"}
var ErrmsgKeys = []string{"errmsg"}
var ResultKeys = []string{"result"}
var ObjectKeys = []string{"object"}
var InputKeys = []string{"input"}
var OutputKeys = []string{"output"}

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
		return errors.New("Not Found In Logic")
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
			if v.CanInterface() && !v.IsNil() {
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

type TaskLogic struct {
	Name    string
	Options map[string]interface{}
	Fail    ILogic
	Done    ILogic
}

func (L *TaskLogic) Exec(a app.IApp, program IProgram, ctx IContext) error {

	task, ok := app.NewTask(a, strings.Split(L.Name, "."))

	if ok {

		v := reflect.ValueOf(task)

		if L.Options != nil {

			for key, value := range L.Options {
				vv := ctx.ReflectValue(value)
				Value.SetWithKeys(v, []string{key}, reflect.ValueOf(vv))
			}

		}

		err := app.Handle(a, task)

		if err != nil {
			if L.Fail != nil {
				ctx.Set(ErrnoKeys, ERROR_UNKNOWN)
				ctx.Set(ErrmsgKeys, err)
				return L.Fail.Exec(a, program, ctx)
			} else {
				return err
			}
		} else {

			rs := task.GetResult()
			rsv := reflect.ValueOf(rs)
			errno := Value.IntValue(Value.GetWithKeys(rsv, []string{"Errno"}), 0)
			errmsg := Value.StringValue(Value.GetWithKeys(rsv, []string{"Errmsg"}), "")

			if errno != 0 {

				if L.Fail != nil {
					ctx.Set(ErrnoKeys, errno)
					ctx.Set(ErrmsgKeys, errmsg)
					return L.Fail.Exec(a, program, ctx)
				} else {
					return errors.New(errmsg)
				}

			} else {

				ctx.Set(ResultKeys, rs)

				if L.Done != nil {
					return L.Done.Exec(a, program, ctx)
				}

				return nil
			}

		}

	} else if L.Fail != nil {
		ctx.Set(ErrnoKeys, ERROR_UNKNOWN)
		ctx.Set(ErrmsgKeys, "Not Found Task "+L.Name)
		return L.Fail.Exec(a, program, ctx)
	} else {
		return errors.New("Not Found Task " + L.Name)
	}
}

type OutputField struct {
	Name  string
	Keys  string
	Value interface{}
	Done  ILogic
}

type OutputLogic struct {
	Keys   string
	Value  interface{}
	Fields []OutputField
	Done   ILogic
}

func toObject(a app.IApp, program IProgram, ctx IContext, value reflect.Value, object map[string]interface{}, fields []OutputField) error {

	for _, fd := range fields {

		if fd.Done != nil {

			ctx.Begin()
			ctx.Set(OutputKeys, map[string]interface{}{})

			if fd.Keys != "" {
				v := Value.GetWithKeys(value, strings.Split(fd.Keys, "."))
				if v.CanInterface() && !v.IsNil() {
					ctx.Set(ObjectKeys, v.Interface())
				}
			}

			err := fd.Done.Exec(a, program, ctx)

			if err != nil {
				ctx.End()
				return err
			}

			object[fd.Name] = ctx.Get(OutputKeys)
			ctx.End()

		} else if fd.Value != nil {
			object[fd.Name] = ctx.ReflectValue(fd.Value)
		} else if fd.Keys != "" {
			v := Value.GetWithKeys(value, strings.Split(fd.Keys, "."))
			if v.CanInterface() && !v.IsNil() {
				object[fd.Name] = v.Interface()
			}
		}
	}

	return nil
}

func (L *OutputLogic) Exec(a app.IApp, program IProgram, ctx IContext) error {

	vv := ctx.ReflectValue(L.Value)
	output := ctx.Get(OutputKeys)

	if output == nil {
		output = map[string]interface{}{}
		ctx.Set(OutputKeys, output)
	}

	keys := strings.Split(L.Keys, ".")

	if L.Fields != nil {

		v := reflect.ValueOf(vv)

		if v.Kind() == reflect.Slice {

			var out = []map[string]interface{}{}

			for i := 0; i < v.Len(); i++ {
				var object = map[string]interface{}{}
				err := toObject(a, program, ctx, v.Index(i), object, L.Fields)
				if err != nil {
					return err
				}
				out = append(out, object)
			}

			Value.SetWithKeys(reflect.ValueOf(output), keys, reflect.ValueOf(out))

		} else {

			var object = map[string]interface{}{}

			err := toObject(a, program, ctx, v, object, L.Fields)

			if err != nil {
				return err
			}

			Value.SetWithKeys(reflect.ValueOf(output), keys, reflect.ValueOf(object))

		}

	} else {
		Value.SetWithKeys(reflect.ValueOf(output), keys, reflect.ValueOf(vv))
	}

	if L.Done != nil {
		return L.Done.Exec(a, program, ctx)
	}

	return nil
}
