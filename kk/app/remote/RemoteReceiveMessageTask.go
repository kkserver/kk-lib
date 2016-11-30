package remote

import (
	"github.com/kkserver/kk-lib/kk"
	"github.com/kkserver/kk-lib/kk/app"
)

type RemoteReceiveMessageTask struct {
	app.Task
	Name    string
	Message kk.Message
}
