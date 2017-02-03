package app

import (
	"database/sql"
	"github.com/kkserver/kk-lib/kk"
	"reflect"
)

type DBConfig struct {
	Name         string
	Url          string
	Prefix       string
	MaxIdleConns int
	MaxOpenConns int
	db           *sql.DB
}

func (C *DBConfig) Get(app IApp) (*sql.DB, error) {

	if C.db == nil {

		var db, err = sql.Open(C.Name, C.Url)

		if err != nil {
			return nil, err
		}

		db.SetMaxIdleConns(C.MaxIdleConns)
		db.SetMaxOpenConns(C.MaxOpenConns)

		err = kk.DBInit(db)

		if err != nil {
			db.Close()
			return nil, err
		}

		var v = reflect.ValueOf(app)

		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		for i := 0; i < v.NumField(); i++ {

			fd := v.Field(i)

			if fd.Kind() == reflect.Struct && fd.CanAddr() {

				r, ok := fd.Addr().Interface().(*kk.DBTable)

				if ok {
					err = kk.DBBuild(db, r, C.Prefix, 1)
					if err != nil {
						db.Close()
						db = nil
						return nil, err
					}
				}

			}

		}

		C.db = db
	}

	return C.db, nil
}

func (C *DBConfig) Recycle() {

	if C.db != nil {
		C.db.Close()
		C.db = nil
	}

}
