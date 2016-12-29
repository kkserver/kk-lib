package logic

import (
	"github.com/kkserver/kk-lib/kk/app"
	Value "github.com/kkserver/kk-lib/kk/value"
	"reflect"
	"strings"
)

type OutputField struct {
	Name  string
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
			ctx.Set(ObjectKeys, value)

			err := fd.Done.Exec(a, program, ctx)

			if err != nil {
				ctx.End()
				return err
			}

			object[fd.Name] = ctx.Get(OutputKeys)
			ctx.End()

		} else if fd.Value != nil {
			ctx.Begin()
			ctx.Set(ObjectKeys, value)
			object[fd.Name] = ctx.ReflectValue(fd.Value)
			ctx.End()
		} else {
			v := Value.GetWithKeys(value, strings.Split(fd.Name, "."))
			if v.IsValid() && v.CanInterface() && !v.IsNil() {
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
