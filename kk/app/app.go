package app

import (
	"errors"
	"fmt"
	"github.com/kkserver/kk-lib/kk/dynamic"
	"github.com/kkserver/kk-lib/kk/inifile"
	"reflect"
	"strings"
)

const ERROR_UNKNOWN = 0x1000

var Break = errors.New("break")

type Error struct {
	Errno  int    `json:"errno,omitempty"`
	Errmsg string `json:"errmsg,omitempty"`
}

func (E *Error) Error() string {
	return fmt.Sprintf("[%d] %s", E.Errno, E.Errmsg)
}

func NewError(errno int, errmsg string) *Error {
	return &Error{errno, errmsg}
}

type Result struct {
	Error
}

type IObtain interface {
	Obtain()
}

type IRecycle interface {
	Recycle()
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
	IObtain
	IRecycle
}

type App struct {
}

func (A *App) Recycle() {

}

func (A *App) Obtain() {

}

func (T *Task) GetResult() interface{} {
	return nil
}

func (S *Service) Handle(app IApp, task ITask) error {
	return ServiceReflectHandle(app, task, S)
}

func ServiceReflectHandle(app IApp, task ITask, service IService) error {
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

		dynamic.SetWithKeys(app, append(keys, f.Key), f.Value)

	}

	return nil
}

func Obtain(obtain IObtain) {

	obtain.Obtain()

	var v = reflect.ValueOf(obtain)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {

		fd := v.Field(i)

		if fd.Kind() == reflect.Ptr && !fd.IsNil() && fd.CanInterface() {

			vv := fd.Interface()

			if vv != nil {

				r, ok := vv.(IObtain)

				if ok {
					Obtain(r)
				}
			}

		}

	}

}

func Recycle(recycle IRecycle) {

	recycle.Recycle()

	var v = reflect.ValueOf(recycle)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < v.NumField(); i++ {

		fd := v.Field(i)

		if fd.Kind() == reflect.Ptr && !fd.IsNil() && fd.CanInterface() {

			vv := fd.Interface()

			if vv != nil {

				r, ok := vv.(IRecycle)

				if ok {
					Recycle(r)
				}
			}

		}

	}

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

	var v = dynamic.GetWithKeys(app, name)

	if v != nil {
		var t, ok = v.(ITask)
		return t, ok
	}

	return nil, false

}
