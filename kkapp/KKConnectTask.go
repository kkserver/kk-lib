package kkapp

import (
	"github.com/kkserver/kk-lib/app"
	"time"
)

type KKConnectTask struct {
	app.Task
	Name    string
	Address string
	Options map[string]interface{}
	Timeout time.Duration
}
