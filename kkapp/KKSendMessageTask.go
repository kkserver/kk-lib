package kkapp

import (
	"github.com/kkserver/kk-lib/app"
	"github.com/kkserver/kk-lib/kk"
)

type KKSendMessageTask struct {
	app.Task
	Message kk.Message
}
