package task

import (
	"reflect"
)

type Context struct {
	services map[reflect.Type][]IService
	value    map[string]interface{}
	apitypes map[string]reflect.Type
}

func NewContext() *Context {
	var v = Context{map[reflect.Type][]IService{}, map[string]interface{}{}, map[string]reflect.Type{}}
	return &v
}

func (C *Context) Get(key string) interface{} {
	var v, ok = C.value[key]
	if ok {
		return v
	}
	return nil
}

func (C *Context) Set(key string, value interface{}) *Context {
	C.value[key] = value
	return C
}

func (C *Context) Handle(task ITask) error {

	var taskType = reflect.TypeOf(task)

	for key, services := range C.services {
		if key == taskType || (key.Kind() == reflect.Interface && taskType.Implements(key)) {
			for _, service := range services {
				var err = service.Handle(task)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (C *Context) Add(task ITask, service IService) {

	var taskType = reflect.TypeOf(task)

	var api, ok = task.(IAPITask)

	if ok {
		C.apitypes[api.API()] = taskType.Elem()
	}

	services, ok := C.services[taskType]

	if !ok {
		services = []IService{service}
		C.services[taskType] = services
	} else {
		C.services[taskType] = append(services, service)
	}

}

func (C *Context) Service(service IService) func(tasks ...ITask) {
	return func(tasks ...ITask) {
		for _, task := range tasks {
			C.Add(task, service)
		}
	}
}

func (C *Context) Plugin(plugin IPlugin) func(service IService) func(tasks ...ITask) {
	return func(service IService) func(tasks ...ITask) {
		service.SetPlugin(plugin)
		return C.Service(service)
	}
}

func (C *Context) NewAPITask(api string) IAPITask {
	var t, ok = C.apitypes[api]
	if ok {
		return reflect.New(t).Interface().(IAPITask)
	}
	return nil
}
