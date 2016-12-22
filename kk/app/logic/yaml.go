package logic

import (
	"bufio"
	"github.com/go-yaml/yaml"
	Value "github.com/kkserver/kk-lib/kk/value"
	"io"
	"os"
	"reflect"
)

type YamlProgram struct {
	Program
	config     map[string]interface{}
	logicTypes map[string]reflect.Type
	logics     map[string]ILogic
}

func NewYamlProgram(path string) (*YamlProgram, error) {

	fd, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer fd.Close()

	rd := bufio.NewReader(fd)

	config := map[string]interface{}{}

	data, err := rd.ReadBytes(0)

	if err != nil && err != io.EOF {
		return nil, err
	}

	err = yaml.Unmarshal(data, config)

	if err != nil {
		return nil, err
	}

	v := YamlProgram{}
	v.config = config
	v.logics = map[string]ILogic{}
	v.logicTypes = map[string]reflect.Type{"Task": reflect.TypeOf(TaskLogic{}), "Output": reflect.TypeOf(OutputLogic{})}
	return &v, nil
}

func (P *YamlProgram) Use(name string, logic ILogic) {
	P.logicTypes[name] = reflect.TypeOf(logic).Elem()
}

func (P *YamlProgram) newLogic(value reflect.Value) ILogic {

	if value.Kind() == reflect.Map {

		name := Value.StringValue(Value.GetWithKeys(value, []string{"logic"}), "")

		if name != "" {

			t, ok := P.logicTypes[name]

			if ok {

				logicV := reflect.New(t)
				logic, ok := logicV.Interface().(ILogic)

				if ok {

					Value.EachObject(value, func(key reflect.Value, value reflect.Value) bool {

						skey := Value.StringValue(key, "")

						if len(skey) > 0 && skey[0] >= 'A' && skey[0] <= 'Z' {

							vv := Value.GetWithKeys(logicV, []string{skey})

							if vv.Kind() == reflect.Interface && vv.Type().String() == "logic.ILogic" {
								lv := P.newLogic(value)
								if lv != nil {
									Value.SetValue(vv, reflect.ValueOf(lv))
								}
								return true
							}

							Value.SetValue(vv, value)
						}

						return true
					})

					return logic
				}
			}

		}

	} else {
		return P.GetLogic(Value.StringValue(value, ""))
	}

	return nil
}

func (P *YamlProgram) GetLogic(name string) ILogic {

	{
		v, ok := P.logics[name]

		if ok {
			return v
		}
	}

	{
		v, ok := P.config[name]

		if !ok {
			return nil
		}

		logic := P.newLogic(reflect.ValueOf(v))

		if logic != nil {
			P.logics[name] = logic
			return logic
		}

	}

	return nil
}
