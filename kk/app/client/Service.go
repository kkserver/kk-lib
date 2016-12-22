package client

import (
	"errors"
	"github.com/kkserver/kk-lib/kk"
	"github.com/kkserver/kk-lib/kk/app"
	"github.com/kkserver/kk-lib/kk/json"
	"time"
)

type Config struct {
	Name    string
	Address string
	Options map[string]interface{}
}

type Service struct {
	app.Service
	RequestTask *RequestTask
	Config      Config

	request func(message *kk.Message, trackId string, timeout time.Duration) *kk.Message
	getName func() string
}

func (S *Service) Handle(a app.IApp, task app.ITask) error {
	return app.ServiceReflectHandle(a, task, S)
}

func (S *Service) HandleInitTask(a app.IApp, task *app.InitTask) error {

	S.request, S.getName, _ = kk.TCPClientRequestConnect(S.Config.Name, S.Config.Address, S.Config.Options)

	return nil
}

func (S *Service) HandleRequestTask(a app.IApp, task *RequestTask) error {

	if S.request != nil {

		var v = kk.Message{}

		v.To = task.Name
		v.Type = "text/json"

		if task.RequestTask != nil {
			v.Content, _ = json.Encode(task.RequestTask)
		} else {
			v.Content, _ = json.Encode(task.Request)
		}

		var r = S.request(&v, task.TrackId, task.Timeout)

		if r == nil {
			return errors.New("client.Service request fail")
		} else if r.Method == "REQUEST" {
			if task.RequestTask != nil {
				if r.Type == "text/json" || r.Type == "application/json" {
					return json.Decode(r.Content, task.RequestTask.GetResult())
				} else {
					return errors.New("client.Service request fail " + r.String())
				}
			} else {
				if r.Type == "text/json" || r.Type == "application/json" {
					return json.Decode(r.Content, &task.Result)
				} else if r.Type == "text" {
					task.Result = string(r.Content)
				}
			}
		} else {
			return errors.New("client.Service request fail " + r.String())
		}

	} else {
		return errors.New("client.Service not connected")
	}

	return nil
}
