package client

import (
	"github.com/kkserver/kk-lib/kk/app"
	"time"
)

type RequestTask struct {
	app.Task
	Name        string
	RequestTask app.ITask
	Timeout     time.Duration
}
