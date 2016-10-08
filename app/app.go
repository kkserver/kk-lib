package app

import (
	"reflect"
)

type IApp interface {
	Parent() IApp
	Handle(task ITask) error
	Get(key string) interface{}
	NewAPITask(api string) IAPITask
}

type App struct {
	parent   IApp
	services map[reflect.Type][]IService
	value    map[string]interface{}
	apitypes map[string]reflect.Type
}

func NewApp(parent IApp) *App {
	var v = App{parent, map[reflect.Type][]IService{}, map[string]interface{}{}, map[string]reflect.Type{}}
	return &v
}

func (app *App) Get(key string) interface{} {
	var v, ok = app.value[key]
	if ok {
		return v
	}
	return nil
}

func (app *App) Set(key string, value interface{}) *App {
	app.value[key] = value
	return app
}

func (app *App) Parent() IApp {
	return app.parent
}

func (app *App) Handle(task ITask) error {

	var taskType = reflect.TypeOf(task)
	var count = 0

	for key, services := range app.services {
		if key == taskType || (key.Kind() == reflect.Interface && taskType.Implements(key)) {
			for _, service := range services {
				var err = service.Handle(app, task)
				count = count + 1
				if err != nil {
					return err
				}
			}
		}
	}

	if count == 0 && app.parent != nil {
		return app.parent.Handle(task)
	}

	return nil
}

func (app *App) Add(task ITask, service IService) *App {

	var taskType = reflect.TypeOf(task)

	var api, ok = task.(IAPITask)

	if ok {
		app.apitypes[api.API()] = taskType.Elem()
	}

	services, ok := app.services[taskType]

	if !ok {
		services = []IService{service}
		app.services[taskType] = services
	} else {
		app.services[taskType] = append(services, service)
	}

	return app
}

func (app *App) Service(service IService) func(tasks ...ITask) {
	return func(tasks ...ITask) {
		for _, task := range tasks {
			app.Add(task, service)
		}
	}
}

func (app *App) NewAPITask(api string) IAPITask {
	var t, ok = app.apitypes[api]
	if ok {
		return reflect.New(t).Interface().(IAPITask)
	}
	if app.parent != nil {
		return app.parent.NewAPITask(api)
	}
	return nil
}
