package logic

import (
	"github.com/kkserver/kk-lib/kk/app"
	"github.com/kkserver/kk-lib/kk/app/client"
	Value "github.com/kkserver/kk-lib/kk/value"
	"reflect"
	"time"
)

type RequestLogic struct {
	Name    string
	Options map[string]interface{}
	Timeout int64
	Fail    ILogic
	Done    ILogic
}

func (L *RequestLogic) Exec(a app.IApp, program IProgram, ctx IContext) error {

	task := client.RequestTask{}

	task.Name = L.Name

	if L.Timeout == 0 {
		task.Timeout = 1 * time.Second
	} else {
		task.Timeout = time.Duration(L.Timeout) * time.Second
	}

	v := map[string]interface{}{}

	task.Request = v

	if L.Options != nil {

		for key, value := range L.Options {
			vv := ctx.ReflectValue(value)
			if key == "_" {
				Value.EachObject(reflect.ValueOf(vv), func(key reflect.Value, vv reflect.Value) bool {
					if vv.IsValid() && vv.CanInterface() && !vv.IsNil() {
						v[Value.StringValue(key, "")] = vv.Interface()
					}
					return true
				})
			} else {
				v[key] = vv
			}

		}

	}

	err := app.Handle(a, &task)

	if err != nil {
		if L.Fail != nil {
			ctx.Set(ErrnoKeys, ERROR_UNKNOWN)
			ctx.Set(ErrmsgKeys, err)
			return L.Fail.Exec(a, program, ctx)
		} else {
			return err
		}
	} else {

		rs := task.Result
		rsv := reflect.ValueOf(rs)
		errno := Value.IntValue(Value.GetWithKeys(rsv, []string{"errno"}), 0)
		errmsg := Value.StringValue(Value.GetWithKeys(rsv, []string{"errmsg"}), "")

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

}
