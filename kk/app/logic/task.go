package logic

import (
	"github.com/kkserver/kk-lib/kk/app"
	Value "github.com/kkserver/kk-lib/kk/value"
	"reflect"
	"strings"
)

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
				if key == "_" {
					Value.EachObject(reflect.ValueOf(vv), func(key reflect.Value, value reflect.Value) bool {
						Value.SetWithKeys(v, []string{Value.StringValue(key, "")}, value)
						return true
					})
				} else {
					Value.SetWithKeys(v, []string{key}, reflect.ValueOf(vv))
				}

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
					return app.NewError(ERROR_UNKNOWN, errmsg)
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
		return app.NewError(ERROR_UNKNOWN, "Not Found Task "+L.Name)
	}
}
