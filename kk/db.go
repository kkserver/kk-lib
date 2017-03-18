package kk

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/kkserver/kk-lib/kk/dynamic"
	"log"
	"reflect"
	"strings"
)

const DBFieldTypeString = 1
const DBFieldTypeInt = 2
const DBFieldTypeInt64 = 3
const DBFieldTypeDouble = 4
const DBFieldTypeBoolean = 5
const DBFieldTypeText = 6
const DBFieldTypeLongText = 7

type DBField struct {
	Length int
	Type   int
}

func (fd *DBField) SetValue(key string, value interface{}) {
	if key == "Length" {
		fd.Length = int(dynamic.IntValue(value, 0))
	} else if key == "Type" {
		switch dynamic.StringValue(value, "") {
		case "string":
			fd.Type = DBFieldTypeString
		case "int":
			fd.Type = DBFieldTypeInt
		case "int64":
			fd.Type = DBFieldTypeInt64
		case "double":
			fd.Type = DBFieldTypeDouble
		case "boolean":
			fd.Type = DBFieldTypeBoolean
		case "text":
			fd.Type = DBFieldTypeText
		case "longtext":
			fd.Type = DBFieldTypeLongText
		default:
			fd.Type = DBFieldTypeString
		}
	}
}

func (fd *DBField) DBType() string {
	switch fd.Type {
	case DBFieldTypeInt:
		if fd.Length == 0 {
			return "INT"
		}
		return fmt.Sprintf("INT(%d)", fd.Length)
	case DBFieldTypeInt64:
		if fd.Length == 0 {
			return "BIGINT"
		}
		return fmt.Sprintf("BIGINT(%d)", fd.Length)
	case DBFieldTypeDouble:
		if fd.Length == 0 {
			return "DOUBLE"
		}
		return fmt.Sprintf("DOUBLE(%d)", fd.Length)
	case DBFieldTypeBoolean:
		return "INT(1)"
	case DBFieldTypeText:
		if fd.Length == 0 {
			return "TEXT"
		}
		return fmt.Sprintf("TEXT(%d)", fd.Length)
	case DBFieldTypeLongText:
		if fd.Length == 0 {
			return "LONGTEXT"
		}
		return fmt.Sprintf("LONGTEXT(%d)", fd.Length)
	}
	if fd.Length == 0 {
		return "VARCHAR(45)"
	}
	return fmt.Sprintf("VARCHAR(%d)", fd.Length)
}

func (fd *DBField) DBDefaultValue() string {
	switch fd.Type {
	case DBFieldTypeInt, DBFieldTypeInt64, DBFieldTypeDouble, DBFieldTypeBoolean:
		return "DEFAULT 0"
	}
	return "DEFAULT ''"
}

func (fd *DBField) String() string {
	return fd.DBType()
}

const DBIndexTypeAsc = 1
const DBIndexTypeDesc = 2

type DBIndex struct {
	Field  string
	Type   int
	Unique bool
}

func (fd *DBIndex) SetValue(key string, value interface{}) {
	switch key {
	case "Field":
		fd.Field = dynamic.StringValue(value, "")
	case "Type":
		switch dynamic.StringValue(value, "") {
		case "string":
			fd.Type = DBFieldTypeString
		case "int":
			fd.Type = DBFieldTypeInt
		case "int64":
			fd.Type = DBFieldTypeInt64
		case "double":
			fd.Type = DBFieldTypeDouble
		case "boolean":
			fd.Type = DBFieldTypeBoolean
		case "text":
			fd.Type = DBFieldTypeText
		case "longtext":
			fd.Type = DBFieldTypeLongText
		default:
			fd.Type = DBFieldTypeString
		}
	case "Unique":
		fd.Unique = dynamic.BooleanValue(value, false)
	}
}

func (idx *DBIndex) DBType() string {
	switch idx.Type {
	case DBIndexTypeAsc:
		return "ASC"
	case DBIndexTypeDesc:
		return "DESC"
	}
	return "ASC"
}

func (idx *DBIndex) String() string {
	return idx.DBType()
}

type DBTable struct {
	Name   string
	Key    string
	Fields map[string]*DBField
	Indexs map[string]*DBIndex
}

func DBInit(db *sql.DB) error {
	var _, err = db.Exec("CREATE TABLE IF NOT EXISTS __scheme (id BIGINT NOT NULL AUTO_INCREMENT,name VARCHAR(64) NULL,scheme TEXT NULL,PRIMARY KEY (id),INDEX name (name ASC) ) AUTO_INCREMENT=1;")
	return err
}

type Database interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func DBBuild(db Database, table *DBTable, prefix string, auto_increment int) error {

	var tbname = prefix + table.Name

	var rs, err = db.Query("SELECT * FROM __scheme WHERE name=?", tbname)

	if err != nil {
		return err
	}

	defer rs.Close()

	if rs.Next() {

		var id int64
		var name string
		var scheme string
		rs.Scan(&id, &name, &scheme)
		var tb DBTable
		json.Unmarshal([]byte(scheme), &tb)
		var hasUpdate = false

		for name, field := range table.Fields {
			var fd, ok = tb.Fields[name]
			if ok {
				if fd.Type != field.Type || fd.Length != field.Length {
					log.Println("SQL", fmt.Sprintf("ALTER TABLE `%s` CHANGE `%s` `%s` %s %s;", tbname, name, name, field.DBType(), field.DBDefaultValue()))
					_, err = db.Exec(fmt.Sprintf("ALTER TABLE `%s` CHANGE `%s` `%s` %s %s;", tbname, name, name, field.DBType(), field.DBDefaultValue()))
					if err != nil {
						log.Println(err)
					}
					hasUpdate = true
				}
			} else {
				log.Println("SQL", fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s %s;", tbname, name, field.DBType(), field.DBDefaultValue()))
				_, err = db.Exec(fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s %s;", tbname, name, field.DBType(), field.DBDefaultValue()))
				if err != nil {
					log.Println(err)
				}
				hasUpdate = true
			}
		}

		for name, index := range table.Indexs {
			var _, ok = tb.Indexs[name]
			if !ok {
				if index.Unique {
					_, err = db.Exec(fmt.Sprintf("CREATE UNIQUE INDEX `%s` ON `%s` (`%s` %s);", name, tbname, name, index.DBType()))
					if err != nil {
						log.Println(err)
					}
				} else {
					_, err = db.Exec(fmt.Sprintf("CREATE INDEX `%s` ON `%s` (`%s` %s);", name, tbname, name, index.DBType()))
					if err != nil {
						log.Println(err)
					}
				}

				hasUpdate = true
			}
		}

		if hasUpdate {
			var b, _ = json.Marshal(table)
			_, err = db.Exec("UPDATE __scheme SET scheme=? WHERE id=?", string(b), id)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {

		var s bytes.Buffer
		var i int = 0

		s.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (", tbname))

		if table.Key != "" {
			s.WriteString(fmt.Sprintf("`%s` BIGINT NOT NULL AUTO_INCREMENT", table.Key))
			i += 1
		}

		for name, field := range table.Fields {
			if i != 0 {
				s.WriteString(",")
			}
			s.WriteString(fmt.Sprintf("`%s` %s %s", name, field.DBType(), field.DBDefaultValue()))
			i += 1
		}

		if table.Key != "" {
			s.WriteString(fmt.Sprintf(", PRIMARY KEY(`%s`) ", table.Key))
		}

		for name, index := range table.Indexs {

			if index.Unique {
				s.WriteString(fmt.Sprintf(",UNIQUE INDEX `%s` (`%s` %s)", name, name, index.DBType()))
			} else {
				s.WriteString(fmt.Sprintf(",INDEX `%s` (`%s` %s)", name, name, index.DBType()))
			}

		}

		if table.Key != "" {
			s.WriteString(fmt.Sprintf(" ) AUTO_INCREMENT = %d;", auto_increment))
		} else {
			s.WriteString(" ) ;")
		}

		log.Println(s.String())

		_, err = db.Exec(s.String())
		if err != nil {
			log.Fatal(err)
		}

		var b, _ = json.Marshal(table)

		_, err = db.Exec("INSERT INTO __scheme(name,scheme) VALUES(?,?)", tbname, string(b))
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func DBQuery(db Database, table *DBTable, prefix string, sql string, args ...interface{}) (*sql.Rows, error) {
	var tbname = prefix + table.Name
	return db.Query(fmt.Sprintf("SELECT * FROM `%s` %s", tbname, sql), args...)
}

func DBQueryWithKeys(db Database, table *DBTable, prefix string, keys map[string]bool, sql string, args ...interface{}) (*sql.Rows, error) {

	s := bytes.NewBuffer(nil)

	if keys == nil {
		s.WriteString("SELECT *")
	} else {
		s.WriteString("SELECT ")
		i := 0
		if table.Key != "" {
			s.WriteString(fmt.Sprintf("`%s`", table.Key))
			i = i + 1
		}
		for name, _ := range table.Fields {
			v, ok := keys[name]
			if ok && v {
				if i != 0 {
					s.WriteString(",")
				}
				s.WriteString(fmt.Sprintf("`%s`", name))
				i = i + 1
			}
		}
	}

	s.WriteString(fmt.Sprintf("FROM `%s%s` %s", prefix, table.Name, sql))

	return db.Query(s.String(), args...)
}

func DBDelete(db Database, table *DBTable, prefix string, sql string, args ...interface{}) (sql.Result, error) {
	var tbname = prefix + table.Name
	return db.Exec(fmt.Sprintf("DELETE FROM `%s` %s", tbname, sql), args...)
}

func DBQueryCount(db Database, table *DBTable, prefix string, sql string, args ...interface{}) (int, error) {
	var tbname = prefix + table.Name

	var rows, err = db.Query(fmt.Sprintf("SELECT COUNT(*) as c FROM `%s` %s", tbname, sql), args...)

	if err != nil {
		return 0, err
	}

	defer rows.Close()

	if rows.Next() {
		var v int = 0
		err = rows.Scan(&v)
		if err != nil {
			return 0, err
		}
		return v, nil
	}

	return 0, nil
}

func DBUpdate(db Database, table *DBTable, prefix string, object interface{}) (sql.Result, error) {
	return DBUpdateWithKeys(db, table, prefix, object, nil)
}

func DBUpdateWithKeys(db Database, table *DBTable, prefix string, object interface{}, keys map[string]bool) (sql.Result, error) {

	var tbname = prefix + table.Name
	var s bytes.Buffer
	var fsc = len(table.Fields)
	var fs = make([]interface{}, fsc+1)
	var key interface{} = nil
	var n = 0

	s.WriteString(fmt.Sprintf("UPDATE `%s` SET ", tbname))

	var tbv = reflect.ValueOf(object).Elem()
	var tb = tbv.Type()
	var fc = tb.NumField()

	for i := 0; i < fc; i += 1 {
		var fd = tb.Field(i)
		var fbv = tbv.Field(i)
		var name = strings.ToLower(fd.Name)
		if name == table.Key {
			key = fbv.Interface()
		} else if keys == nil || keys[name] {
			var _, ok = table.Fields[name]
			if ok {
				if n != 0 {
					s.WriteString(",")
				}
				s.WriteString(fmt.Sprintf(" `%s`=?", name))
				fs[n] = fbv.Interface()
				n += 1
			}
		}
	}

	s.WriteString(fmt.Sprintf(" WHERE `%s`=?", table.Key))

	fs[n] = key

	n += 1

	log.Printf("%s %s\n", s.String(), fs)

	return db.Exec(s.String(), fs[:n]...)
}

func DBInsert(db Database, table *DBTable, prefix string, object interface{}) (sql.Result, error) {
	var tbname = prefix + table.Name
	var s bytes.Buffer
	var w bytes.Buffer
	var fsc = len(table.Fields)
	var fs = make([]interface{}, fsc)
	var n = 0
	var key reflect.Value

	s.WriteString(fmt.Sprintf("INSERT INTO `%s`(", tbname))
	w.WriteString(" VALUES (")

	var tbv = reflect.ValueOf(object).Elem()
	var tb = tbv.Type()
	var fc = tb.NumField()

	for i := 0; i < fc; i += 1 {
		var fd = tb.Field(i)
		var fbv = tbv.Field(i)
		var name = strings.ToLower(fd.Name)
		if name == table.Key {
			key = fbv
		} else {
			var _, ok = table.Fields[name]
			if ok {
				if n != 0 {
					s.WriteString(",")
					w.WriteString(",")
				}
				s.WriteString("`" + name + "`")
				w.WriteString("?")
				fs[n] = fbv.Interface()
				n += 1
			}
		}
	}

	s.WriteString(")")

	w.WriteString(")")

	s.Write(w.Bytes())

	log.Printf("%s %s\n", s.String(), fs)

	var rs, err = db.Exec(s.String(), fs[:n]...)

	if err == nil && key.CanSet() {
		id, err := rs.LastInsertId()
		if err == nil {
			key.SetInt(id)
		}
	}

	return rs, err
}

type DBValue struct {
	String  string
	Int64   int64
	Double  float64
	Boolean bool
}

type DBScaner struct {
	object   interface{}
	fields   []interface{}
	nilValue interface{}
}

func NewDBScaner(object interface{}) DBScaner {
	return DBScaner{object, nil, ""}
}

func (o *DBScaner) Scan(rows *sql.Rows) error {

	if o.fields == nil {

		var columns, err = rows.Columns()

		if err != nil {
			return err
		}

		var fdc = len(columns)
		var mi = map[string]int{}

		for i := 0; i < fdc; i += 1 {
			mi[columns[i]] = i
		}

		o.fields = make([]interface{}, fdc)

		for i := 0; i < fdc; i += 1 {
			o.fields[i] = &o.nilValue
		}

		var fn func(value reflect.Value) = nil

		fn = func(value reflect.Value) {
			switch value.Kind() {
			case reflect.Ptr:
				if !value.IsNil() {
					fn(value.Elem())
				}
			case reflect.Struct:
				t := value.Type()
				for i := 0; i < value.NumField(); i++ {
					var fd = value.Field(i)
					switch fd.Kind() {
					case reflect.Struct:
						fn(fd)
					default:
						var name = strings.ToLower(t.Field(i).Name)
						idx, ok := mi[name]
						if ok {
							o.fields[idx] = fd.Addr().Interface()
						}
					}

				}
			}
		}

		fn(reflect.ValueOf(o.object))

	}

	return rows.Scan(o.fields...)
}
