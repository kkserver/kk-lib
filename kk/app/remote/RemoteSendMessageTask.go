package remote

import (
	app "../"
	kk "../../"
)

type RemoteSendMessageTask struct {
	app.Task
	Message kk.Message
}
