package main

import (
	"github.com/kkserver/kk-lib/kk"
	"github.com/kkserver/kk-lib/kk/app"
	"github.com/kkserver/kk-lib/kk/app/logic"
	"log"
)

type MainApp struct {
	app.App
}

func main() {

	log.SetFlags(log.Llongfile | log.LstdFlags)

	a := MainApp{}

	err := app.Load(&a, "./app.ini")

	if err != nil {
		log.Panicln(err)
	}

	app.Handle(&a, &app.InitTask{})

	program, err := logic.NewYamlProgram("./demo.yaml")

	if err != nil {
		log.Panic(err)
	}

	ctx := logic.NewLuaContext()

	defer ctx.Close()

	ctx.Begin()

	ctx.Set(logic.OutputKeys, map[string]interface{}{})

	err = logic.Exec(&a, program, ctx)

	if err != nil {
		log.Println(err)
	}

	log.Println(program)

	log.Println(ctx.Get(logic.OutputKeys))

	ctx.End()

	kk.DispatchMain()

}
