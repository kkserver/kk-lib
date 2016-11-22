package app

import (
	"../inifile"
	"../value"
	"errors"
	"reflect"
	"strings"
)

const ERROR_UNKNOWN = 0x1000

var Break = errors.New("break")

type Result struct {
	Errno  int    `json:"errno"`
	Errmsg string `json:"errmsg"`
}

type ITask interface {
	GetResult() interface{}
}

type Task struct {
}

type IService interface {
	Handle(app IApp, task ITask) error
}

type Service struct {
}

type IApp interface {
}

type App struct {
}

func (T *Task) GetResult() interface{} {
	return nil
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

func Load(app IApp, path string) error {

	f, err := inifile.Open(path)

	if err != nil {
		return err
	}

	defer f.Close()

	for f.Next() {

		var keys []string

		if f.Section == "" {
			keys = []string{}
		} else {
			keys = strings.Split(f.Section, ".")
		}

		keys = append(keys, f.Key)

		var v = reflect.ValueOf(app)

		for _, key := range keys {
			v = value.Get(v, key)
			switch v.Kind() {
			case reflect.Interface:
				if v.IsNil() {
					v.Set(reflect.ValueOf(map[string]interface{}{}))
				}
			case reflect.Ptr:
				if v.IsNil() {
					v.Set(reflect.New(v.Type().Elem()))
				}
			}
		}

		value.SetValue(v, reflect.ValueOf(f.Value))

	}

	return nil
}

func Handle(app IApp, task ITask) error {

	var v = reflect.ValueOf(app)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {

		fd := v.Field(i)

		if fd.Kind() == reflect.Ptr && !fd.IsNil() && fd.CanInterface() {

			vv := fd.Interface()

			if vv != nil {

				s, ok := vv.(IService)

				if ok {

					err := s.Handle(app, task)

					if err != nil {
						return err
					}
				}
			}

		}

	}

	return nil

}

func NewTask(app IApp, name []string) (ITask, bool) {

	var v = value.GetWithKeys(reflect.ValueOf(app), name)

	switch v.Kind() {
	case reflect.Ptr:
		v = reflect.New(v.Type().Elem())
	case reflect.Struct:
		v = reflect.New(v.Type())
	default:
		return nil, false
	}

	if v.CanInterface() {
		var t, ok = v.Interface().(ITask)
		return t, ok
	}

	return nil, false

}
