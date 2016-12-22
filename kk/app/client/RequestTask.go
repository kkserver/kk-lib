package client

import (
	"github.com/kkserver/kk-lib/kk/app"
	"time"
)

type RequestTask struct {
	app.Task
	Name        string
	TrackId     string
	RequestTask app.ITask
	Request     interface{}
	Result      interface{}
	Timeout     time.Duration
}
