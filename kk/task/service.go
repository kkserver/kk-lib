package task

import (
	"reflect"
	"strings"
)

type IService interface {
	Plugin() IPlugin
	SetPlugin(plugin IPlugin) error
	Handle(task ITask) error
}

type Service struct {
	_plugin IPlugin
}

func (S *Service) Handle(task ITask) error {
	return S.ReflectHandle(task, S)
}

func (S *Service) Plugin() IPlugin {
	return S._plugin
}

func (S *Service) SetPlugin(plugin IPlugin) error {
	S._plugin = plugin
	return nil
}

func (S *Service) ReflectHandle(task ITask, service IService) error {
	var t = reflect.TypeOf(task)
	var name = t.String()
	var v = reflect.ValueOf(service)
	var i = strings.LastIndex(name, ".")
	var mt = v.MethodByName("Handle" + name[i+1:])
	if mt.IsValid() {
		var rs = mt.Call([]reflect.Value{reflect.ValueOf(task)})
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
