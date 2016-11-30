package remote

import (
	"github.com/kkserver/kk-lib/kk"
	"github.com/kkserver/kk-lib/kk/app"
)

type RemoteSendMessageTask struct {
	app.Task
	Message kk.Message
}
