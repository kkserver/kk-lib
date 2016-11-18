package main

import (
	"./kk/json"
	"log"
)

type B struct {
	Name string `json:"name"`
}

type A struct {
	Title string `json:"title"`
	Id    int64  `json:"id"`
	B     B      `json:"b"`
}

func main() {

	log.SetFlags(log.Llongfile | log.LstdFlags)

	var v = A{}
	var err = json.Decode([]byte("{\"title\":\"ok\",\"id\":\"123123\",\"b\":{\"name\":121243}}"), &v)
	if err != nil {
		log.Panicln(err)
	}
	log.Println(v)
}
