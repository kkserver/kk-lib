package logic

import (
	"fmt"
	"github.com/aarzilli/golua/lua"
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

	return &v
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

				var vv interface{} = C.L.ToGoStruct(-1)

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
		L.L.Close()
		L.L = nil
	}
}
