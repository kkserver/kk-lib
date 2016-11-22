package remote

import (
	app "../"
	kk "../../"
)

type RemoteReceiveMessageTask struct {
	app.Task
	Name    string
	Message kk.Message
}
