package kkapp

import (
	"encoding/json"
	"github.com/kkserver/kk-lib/app"
	"github.com/kkserver/kk-lib/kk"
	"log"
	"strings"
	"time"
)

type KKService struct {
	app.Service
	client *kk.TCPClient
}

func (S *KKService) Handle(a app.IApp, task app.ITask) error {
	return S.ReflectHandle(a, task, S)
}

func (S *KKService) onMessage(a app.IApp, message *kk.Message) {

	if message.Method != "REQUEST" {
		var v = KKReciveMessageTask{}
		v.Message = *message
		a.Handle(v)
		return
	}

	if strings.HasPrefix(message.To, S.client.Name()) {
		var v = KKSendMessageTask{}
		v.Message = kk.Message{"NOIMPLEMENT", message.To, message.From, "text", []byte(message.To)}
		a.Handle(v)
		return
	}

	var apiname = message.To[len(S.client.Name()):]
	var tk = a.NewAPITask(apiname)

	if tk == nil {
		var v = KKSendMessageTask{}
		v.Message = kk.Message{"NOIMPLEMENT", message.To, message.From, "text", []byte(apiname)}
		a.Handle(v)
		return
	} else if message.Type == "text/json" {
		var err = json.Unmarshal(message.Content, tk)
		if err != nil {
			var b, _ = json.Marshal(&app.Result{app.ERROR_UNKNOWN, err.Error()})
			var v = KKSendMessageTask{}
			v.Message = kk.Message{message.Method, message.To, message.From, "text/json", b}
			a.Handle(v)
			return
		}
	}

	go func() {
		var err = a.Handle(tk)
		if err != nil && err != app.ERROR_BREAK {
			var b, _ = json.Marshal(&app.Result{app.ERROR_UNKNOWN, err.Error()})
			var v = KKSendMessageTask{}
			v.Message = kk.Message{message.Method, message.To, message.From, "text/json", b}
			kk.GetDispatchMain().Async(func() {
				a.Handle(v)
			})
			return
		} else {
			var rs, ok = tk.(app.IAPITask)
			if ok {
				var b, _ = json.Marshal(rs.GetResult())
				var v = KKSendMessageTask{}
				v.Message = kk.Message{message.Method, message.To, message.From, "text/json", b}
				kk.GetDispatchMain().Async(func() {
					a.Handle(v)
				})
			} else {
				var v = KKSendMessageTask{}
				v.Message = kk.Message{message.Method, message.To, message.From, "text/json", []byte("{}")}
				kk.GetDispatchMain().Async(func() {
					a.Handle(v)
				})
			}
		}
	}()
}

func (S *KKService) connect(a app.IApp, name string, address string, options map[string]interface{}, timeout time.Duration) {

	log.Printf("Connect(%s) %s ...\n", name, address)

	var cli = kk.NewTCPClient(name, address, options)

	cli.OnConnected = func() {
		log.Printf("Connected(%s) %s \n", name, cli.Address())
	}

	cli.OnDisconnected = func(err error) {
		log.Printf("Disconnected(%s) %s %s\n", name, cli.Address(), err.Error())
		kk.GetDispatchMain().AsyncDelay(func() {
			S.connect(a, name, address, options, timeout)
		}, timeout)
	}

	cli.OnMessage = func(message *kk.Message) {
		S.onMessage(a, message)
	}

	S.client = cli
}

func (S *KKService) disconnect() {

	if S.client != nil {
		S.client.OnConnected = nil
		S.client.OnDisconnected = nil
		S.client.OnMessage = nil
		S.client.Break()
		S.client = nil
	}

}

func (S *KKService) HandleKKConnectTask(a app.IApp, task *KKConnectTask) error {

	S.disconnect()

	S.connect(a, task.Name, task.Address, task.Options, task.Timeout)

	return nil
}

func (S *KKService) HandleKKDisconnectTask(a app.IApp, task *KKDisconnectTask) error {

	S.disconnect()

	return nil
}

func (S *KKService) HandleKKSendMessageTask(a app.IApp, task *KKSendMessageTask) error {

	if S.client != nil {
		S.client.Send(&task.Message, nil)
	}

	return nil
}
