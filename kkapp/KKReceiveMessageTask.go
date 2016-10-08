package kkapp

import (
	"github.com/kkserver/kk-lib/app"
	"github.com/kkserver/kk-lib/kk"
)

type KKReciveMessageTask struct {
	app.Task
	Message kk.Message
}
