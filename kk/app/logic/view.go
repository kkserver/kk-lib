package logic

import (
	"bufio"
	"bytes"
	"github.com/kkserver/kk-lib/kk/app"
	Value "github.com/kkserver/kk-lib/kk/value"
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"
)

type VarLogic struct {
	Keys  string
	Value interface{}
	Done  ILogic
}

func (L *VarLogic) Exec(a app.IApp, program IProgram, ctx IContext) error {

	vv := ctx.ReflectValue(L.Value)

	keys := strings.Split(L.Keys, ".")

	ctx.Set(keys, vv)

	if L.Done != nil {
		return L.Done.Exec(a, program, ctx)
	}

	return nil
}

type View struct {
	Content     []byte
	ContentType string
}

type ViewLogic struct {
	Path        string
	ContentType string
	Fail        ILogic

	content    string
	hasContent bool
}

var viewLogicCodeRegexp, _ = regexp.Compile("\\{\\#.*?\\#\\}")
var viewLogicIncludeRegexp, _ = regexp.Compile("\\<\\!\\-\\-\\ include\\(.*?\\)\\ \\-\\-\\>")

func (L *ViewLogic) Exec(a app.IApp, program IProgram, ctx IContext) error {
	return L.ExecCode(a, program, ctx, func(code string) string {
		return Value.StringValue(reflect.ValueOf(ctx.Get(strings.Split(code, "."))), "")
	})
}

func GetFileContent(path string) (string, error) {

	fd, err := os.Open(path)

	if err != nil {
		return "", err
	}

	rd := bufio.NewReader(fd)

	v, err := rd.ReadString(0)

	fd.Close()

	if err != nil && err != io.EOF {
		return "", err
	}

	data := bytes.NewBuffer(nil)
	i := 0

	for i < len(v) {

		vs := viewLogicIncludeRegexp.FindStringIndex(v[i:])

		if vs != nil {

			if vs[0] > 0 {
				data.WriteString(v[i : i+vs[0]])
			}

			vv, err := GetFileContent(v[i+vs[0]+13 : i+vs[1]-5])

			if err != nil {
				return "", err
			} else {
				data.WriteString(vv)
			}

			i = i + vs[1]

		} else {
			data.WriteString(v[i:])
			break
		}
	}

	return data.String(), nil
}

func (L *ViewLogic) ExecCode(a app.IApp, program IProgram, ctx IContext, code func(code string) string) error {

	if !L.hasContent {

		v, err := GetFileContent(L.Path)

		if err != nil {
			L.content = err.Error()
		} else {
			L.content = v
		}

		L.hasContent = true
	}

	data := bytes.NewBuffer(nil)

	i := 0

	for i < len(L.content) {

		vs := viewLogicCodeRegexp.FindStringIndex(L.content[i:])

		if vs != nil {

			if vs[0] > 0 {
				data.WriteString(L.content[i : i+vs[0]])
			}

			data.WriteString(code(L.content[i+vs[0]+2 : i+vs[1]-2]))

			i = i + vs[1]

		} else {
			data.WriteString(L.content[i:])
			break
		}
	}

	ctx.Set(ViewKeys, &View{data.Bytes(), L.ContentType})

	return nil
}
