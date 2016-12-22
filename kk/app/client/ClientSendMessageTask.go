package client

import (
	"github.com/kkserver/kk-lib/kk"
	"github.com/kkserver/kk-lib/kk/app"
)

type ClientSendMessageTask struct {
	app.Task
	Message kk.Message
}
