package main

import (
	"./kk"
	"./kk/app"
	"./kk/app/client"
	"./kk/app/remote"
	"log"
)

type MainApp struct {
	app.App
	DB     app.DBConfig
	Remote *remote.Service
	Client *client.Service
}

func main() {

	log.SetFlags(log.Llongfile | log.LstdFlags)

	a := MainApp{}

	err := app.Load(&a, "./app.ini")

	if err != nil {
		log.Panicln(err)
	}

	app.Handle(&a, &app.InitTask{})

	kk.DispatchMain()

}
