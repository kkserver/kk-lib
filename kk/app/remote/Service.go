package remote

import (
	"github.com/kkserver/kk-lib/kk"
	"github.com/kkserver/kk-lib/kk/app"
	"github.com/kkserver/kk-lib/kk/json"
	"log"
	"strings"
	"time"
)

type Config struct {
	Name         string
	Address      string
	Ping         string
	PingInterval int64
	Options      map[string]interface{}
	Timeout      int64
}

type Counter struct {
	Count       int64 `json:"count"`
	Interval    int64 `json:"interval"`
	MinInterval int64 `json:"minInterval"`
	MaxInterval int64 `json:"maxInterval"`
	Atime       int64 `json:"atime"`
	Duration    int64 `json:"duration"`
	Size        int64 `json:"size"`
}

type Service struct {
	app.Service
	SendMessage *RemoteSendMessageTask
	Config      Config

	client  *kk.TCPClient
	address string
	counter Counter
	tasks   map[string]*Counter
	ctime   int64
}

func (S *Service) Handle(a app.IApp, task app.ITask) error {
	return app.ServiceReflectHandle(a, task, S)
}

func (S *Service) HandleInitTask(a app.IApp, task *app.InitTask) error {

	S.ctime = time.Now().Unix()
	S.tasks = map[string]*Counter{}

	S.connect(a)

	if S.Config.Ping != "" {

		if S.Config.PingInterval == 0 {
			S.Config.PingInterval = 6
		}

		var ping func() = nil

		ping = func() {

			var v = RemoteSendMessageTask{}
			v.Message.Method = "PING"
			v.Message.To = S.Config.Ping
			v.Message.Type = "text/json"
			v.Message.Content, _ = json.Encode(map[string]interface{}{"options": S.Config.Options, "address": S.address, "counter": &S.counter, "tasks": S.tasks, "ctime": S.ctime})

			S.HandleRemoteSendMessageTask(a, &v)

			kk.GetDispatchMain().AsyncDelay(ping, time.Duration(S.Config.PingInterval)*time.Second)
		}

		kk.GetDispatchMain().Async(ping)
	}

	return nil
}

func (S *Service) onMessage(a app.IApp, message *kk.Message) {

	if message.Method == "CONNECTED" {
		if message.Type == "text" {
			S.address = string(message.Content)
		}
	}

	if message.Method != "REQUEST" {
		var v = RemoteReceiveMessageTask{}
		v.Message = *message
		v.Name = S.client.Name()
		app.Handle(a, &v)
		return
	}

	if !strings.HasPrefix(message.To, S.client.Name()) {
		var v = RemoteSendMessageTask{}
		v.Message = kk.Message{"NOIMPLEMENT", message.To, message.From, "text", []byte(message.To)}
		S.HandleRemoteSendMessageTask(a, &v)
		return
	}

	var name = message.To[len(S.client.Name()):]
	var tk, ok = app.NewTask(a, strings.Split(name, "."))

	log.Println(name)

	if !ok {
		var v = RemoteSendMessageTask{}
		v.Message = kk.Message{"NOIMPLEMENT", message.To, message.From, "text", []byte(name)}
		S.HandleRemoteSendMessageTask(a, &v)
		return
	} else if message.Type == "text/json" || message.Type == "application/json" {
		var err = json.Decode(message.Content, tk)
		if err != nil {
			var b, _ = json.Encode(app.NewError(app.ERROR_UNKNOWN, "[json.Decode] ["+err.Error()+"] "+string(message.Content)))
			var v = RemoteSendMessageTask{}
			v.Message = kk.Message{message.Method, message.To, message.From, "text/json", b}
			S.HandleRemoteSendMessageTask(a, &v)
			return
		}
	}

	var atime = time.Now().UnixNano()
	var size = int64(len(message.Content))

	go func() {

		var err = app.Handle(a, tk)
		var interval = time.Now().UnixNano() - atime

		kk.GetDispatchMain().Async(func() {

			S.counter.Interval = (S.counter.Count*S.counter.Interval + interval) / (S.counter.Count + 1)
			S.counter.Count = S.counter.Count + 1
			S.counter.Atime = atime
			S.counter.Duration = S.counter.Duration + interval
			S.counter.Size = S.counter.Size + size

			if S.counter.Count == 1 {
				S.counter.MaxInterval = interval
				S.counter.MinInterval = interval
			} else {
				if interval > S.counter.MaxInterval {
					S.counter.MaxInterval = interval
				}
				if interval < S.counter.MinInterval {
					S.counter.MinInterval = interval
				}
			}

			v, ok := S.tasks[name]

			if !ok {
				v = &Counter{}
				S.tasks[name] = v
			}

			v.Interval = (v.Count*v.Interval + interval) / (v.Count + 1)
			v.Count = v.Count + 1
			v.Atime = atime
			v.Duration = v.Duration + interval
			v.Size = v.Size + size

			if v.Count == 1 {
				v.MaxInterval = interval
				v.MinInterval = interval
			} else {
				if interval > v.MaxInterval {
					v.MaxInterval = interval
				}
				if interval < v.MinInterval {
					v.MinInterval = interval
				}
			}

		})

		if err != nil && err != app.Break {
			var b, _ = json.Encode(app.NewError(app.ERROR_UNKNOWN, err.Error()))
			var v = RemoteSendMessageTask{}
			v.Message = kk.Message{message.Method, message.To, message.From, "text/json", b}
			kk.GetDispatchMain().Async(func() {
				S.HandleRemoteSendMessageTask(a, &v)
			})
			return
		} else {
			var b, _ = json.Encode(tk.GetResult())
			var v = RemoteSendMessageTask{}
			v.Message = kk.Message{message.Method, message.To, message.From, "text/json", b}
			kk.GetDispatchMain().Async(func() {
				S.HandleRemoteSendMessageTask(a, &v)
			})
		}
	}()
}

func (S *Service) connect(a app.IApp) {

	log.Printf("Connect(%s) %s ...\n", S.Config.Name, S.Config.Address)

	var cli = kk.NewTCPClient(S.Config.Name, S.Config.Address, S.Config.Options)

	cli.OnConnected = func() {
		log.Printf("Connected(%s) %s \n", S.Config.Name, cli.Address())
	}

	cli.OnDisconnected = func(err error) {
		log.Printf("Disconnected(%s) %s %s\n", S.Config.Name, cli.Address(), err.Error())
		kk.GetDispatchMain().AsyncDelay(func() {
			S.connect(a)
		}, time.Duration(S.Config.Timeout)*time.Second)
	}

	cli.OnMessage = func(message *kk.Message) {
		S.onMessage(a, message)
	}

	S.client = cli
}

func (S *Service) HandleRemoteSendMessageTask(a app.IApp, task *RemoteSendMessageTask) error {

	if S.client != nil {

		if task.Message.From == "" {
			task.Message.From = S.client.Name()
		}

		S.client.Send(&task.Message, nil)
	}

	return nil
}
