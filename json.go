package main

import (
	"./kk/json"
	"log"
)

type B struct {
	Name string `json:"name"`
}

type A struct {
	Title    string `json:"title"`
	Id       int64  `json:"id"`
	Version  int    `json:"version"`
	B        B      `json:"b"`
	Bs       []B    `json:"bs"`
	Booleans []bool `json:"booleans"`
}

func main() {

	log.SetFlags(log.Llongfile | log.LstdFlags)

	var v = A{}
	var err = json.Decode([]byte("{\"title\":\"ok\",\"id\":\"123123\",\"b\":{\"name\":121243},\"version\":\"123\",\"bs\":[{\"name\":1213},{\"name\":123123213213}],\"booleans\":[true,false,123]}"), &v)
	if err != nil {
		log.Panicln(err)
	}
	log.Println(v)
}
