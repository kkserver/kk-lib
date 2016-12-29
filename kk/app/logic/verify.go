package logic

import (
	"github.com/kkserver/kk-lib/kk/app"
	Value "github.com/kkserver/kk-lib/kk/value"
	"reflect"
)

type VerifyLogic struct {
	Value  interface{}
	Errno  interface{}
	Errmsg interface{}
	Fail   ILogic
	Done   ILogic
}

func (L *VerifyLogic) Exec(a app.IApp, program IProgram, ctx IContext) error {

	vv := ctx.ReflectValue(L.Value)

	if Value.BooleanValue(reflect.ValueOf(vv), false) {
		if L.Done != nil {
			return L.Done.Exec(a, program, ctx)
		}
	} else {
		errno := ctx.ReflectValue(L.Errno)
		errmsg := ctx.ReflectValue(L.Errmsg)
		if L.Fail != nil {
			ctx.Set(ErrnoKeys, errno)
			ctx.Set(ErrmsgKeys, errmsg)
			return L.Fail.Exec(a, program, ctx)
		} else {
			return app.NewError(int(Value.IntValue(reflect.ValueOf(errno), 0)), Value.StringValue(reflect.ValueOf(errmsg), "未知错误"))
		}
	}

	return nil
}
