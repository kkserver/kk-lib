package inifile

import (
	"bufio"
	"github.com/kkserver/kk-lib/kk/value"
	"os"
	"reflect"
	"strings"
)

type IniFile struct {
	fd      *os.File
	rd      *bufio.Reader
	Section string
	Key     string
	Value   string
}

func Open(name string) (*IniFile, error) {

	var v = IniFile{}
	var err error = nil

	v.fd, err = os.Open(name)

	if err != nil {
		return nil, err
	}

	v.rd = bufio.NewReader(v.fd)

	return &v, err
}

func (F *IniFile) Close() error {
	return F.fd.Close()
}

func (F *IniFile) Next() bool {

	for {

		line, err := F.rd.ReadSlice('\n')

		if err != nil {
			return false
		}

		sline := strings.TrimSpace(string(line))
		if strings.HasPrefix(sline, "#") {
		} else if strings.HasPrefix(sline, "[") && strings.HasSuffix(sline, "]") {
			F.Section = sline[1 : len(sline)-1]
		} else if strings.Contains(sline, "=") {
			i := strings.Index(sline, "=")
			F.Key = strings.TrimSpace(sline[0:i])
			F.Value = strings.TrimSpace(sline[i+1:])
			return true
		}
	}

	return false
}

func (F *IniFile) Decode(object interface{}) {

	var v = reflect.ValueOf(object)

	for F.Next() {

		var keys []string

		if F.Section == "" {
			keys = []string{}
		} else {
			keys = strings.Split(F.Section, ".")
		}

		value.SetWithKeys(v, append(keys, F.Key), reflect.ValueOf(F.Value))

	}

}

func (F *IniFile) DecodeSection(object interface{}, section string) {

	v := reflect.ValueOf(object)

	for F.Next() {

		if F.Section == section {
			value.Set(v, F.Key, reflect.ValueOf(F.Value))
		}

	}

}

func DecodeFile(object interface{}, path string) error {
	fd, err := Open(path)
	if err != nil {
		return err
	}
	defer fd.Close()
	fd.Decode(object)
	return nil
}
