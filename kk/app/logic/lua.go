package logic

import (
	"fmt"
	"github.com/aarzilli/golua/lua"
	"github.com/kkserver/kk-lib/kk/app"
	"github.com/kkserver/kk-lib/kk/json"
	Value "github.com/kkserver/kk-lib/kk/value"
	"reflect"
	"strings"
)

type LuaContext struct {
	Context
	L *lua.State
}

func NewLuaContext() *LuaContext {

	L := lua.NewState()

	L.OpenLibs()

	v := LuaContext{}
	v.L = L

	L.PushGoFunction(func(L *lua.State) int {

		keys := []string{}
		top := L.GetTop()

		for i := 0; i < top; i++ {
			keys = append(keys, L.ToString(-top+i))
		}

		vv := v.Get(keys)

		L.PushGoStruct(vv)

		return 1
	})

	L.SetGlobal("get")

	L.PushGoFunction(func(L *lua.State) int {

		keys := []string{}
		top := L.GetTop()

		for i := 0; i < top; i++ {
			keys = append(keys, L.ToString(-top+i))
		}

		vv := v.Get(keys)

		b, _ := json.Encode(vv)

		L.PushString(string(b))

		return 1
	})

	L.SetGlobal("json")

	return &v
}

func LuaToValue(L *lua.State, i int) interface{} {

	var vv interface{} = nil

	if L.IsString(i) {
		vv = L.ToString(i)
	} else if L.IsGoStruct(i) {
		vv = L.ToGoStruct(i)
	} else if L.IsNumber(i) {
		vv = L.ToNumber(i)
	} else if L.IsBoolean(i) {
		vv = L.ToBoolean(i)
	} else if L.IsTable(i) {

		L.PushValue(i)

		idx := 0
		size := 0

		m := map[string]interface{}{}
		a := []interface{}{}

		L.PushNil()

		for L.Next(-2) != 0 {

			t := L.Type(-2)

			if t == lua.LUA_TNUMBER {
				if idx == L.ToInteger(-2)-1 {
					a = append(a, LuaToValue(L, -1))
					idx = idx + 1
				}
				m[fmt.Sprintf("%d", L.ToInteger(-2))] = LuaToValue(L, -1)
			} else if t == lua.LUA_TSTRING {
				m[L.ToString(-2)] = LuaToValue(L, -1)
			}

			size = size + 1

			L.Pop(1)

		}

		if idx != 0 && idx == size {
			vv = a
		} else {
			vv = m
		}

		L.Pop(1)
	}

	return vv
}

func (C *LuaContext) ReflectValue(value interface{}) interface{} {

	if C.L == nil {
		return C.Context.ReflectValue(value)
	}

	v := reflect.ValueOf(value)

	if v.Kind() == reflect.String {

		s := v.String()

		if strings.HasPrefix(s, "?lua") {

			if C.L.LoadString(fmt.Sprintf("return %s", s[4:])) == 0 {

				err := C.L.Call(0, 1)

				if err != nil {
					return err.Error()
				}

				if C.L.IsFunction(-1) {

					err = C.L.Call(0, 1)

					if err != nil {
						return err.Error()
					}
				}

				var vv interface{} = LuaToValue(C.L, -1)

				C.L.Pop(1)

				return vv
			} else {
				vv := C.L.ToString(-1)
				C.L.Pop(1)
				return vv
			}

			return C.Get(strings.Split(s[1:], "."))
		}
	}

	return C.Context.ReflectValue(value)
}

func (L *LuaContext) Close() {
	if L.L != nil {
		L.L.PushNil()
		L.L.SetGlobal("get")
		L.L.Close()
		L.L = nil
	}
}

type LuaViewLogic struct {
	ViewLogic
}

func (L *LuaViewLogic) Exec(a app.IApp, program IProgram, ctx IContext) error {
	return L.ExecCode(a, program, ctx, func(code string) string {

		var C *LuaContext = ctx.(*LuaContext)

		var vv interface{} = nil

		if C.L.LoadString(fmt.Sprintf("return %s", code)) == 0 {

			err := C.L.Call(0, 1)

			if err != nil {
				vv = err.Error()
			} else {

				if C.L.IsFunction(-1) {

					err = C.L.Call(0, 1)

					if err != nil {
						vv = err.Error()
					} else {

						vv = LuaToValue(C.L, -1)

						C.L.Pop(1)
					}

				} else {

					vv = LuaToValue(C.L, -1)

					C.L.Pop(1)
				}

			}

		} else {
			vv = C.L.ToString(-1)
			C.L.Pop(1)
		}

		return Value.StringValue(reflect.ValueOf(vv), "")
	})
}
