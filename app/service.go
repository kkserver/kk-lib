package app

import (
	"reflect"
	"strings"
)

type IService interface {
	Handle(app IApp, task ITask) error
}

type Service struct {
}

func (S *Service) Handle(app IApp, task ITask) error {
	return S.ReflectHandle(app, task, S)
}

func (S *Service) ReflectHandle(app IApp, task ITask, service IService) error {
	var t = reflect.TypeOf(task)
	var name = t.String()
	var v = reflect.ValueOf(service)
	var i = strings.LastIndex(name, ".")
	var mt = v.MethodByName("Handle" + name[i+1:])
	if mt.IsValid() {
		var rs = mt.Call([]reflect.Value{reflect.ValueOf(app), reflect.ValueOf(task)})
		if rs != nil && len(rs) > 0 {
			var r = rs[0]
			if r.IsNil() {
				return nil
			} else if r.CanInterface() {
				return r.Interface().(error)
			}
		}
	}
	return nil
}
