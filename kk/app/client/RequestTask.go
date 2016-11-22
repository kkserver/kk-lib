package client

import (
	app "../"
	"time"
)

type RequestTask struct {
	app.Task
	Name        string
	RequestTask app.ITask
	Timeout     time.Duration
}
