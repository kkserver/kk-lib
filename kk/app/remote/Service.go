package remote

import (
	app "../"
	kk "../../"
	json "../../json"
	"log"
	"strings"
	"time"
)

type Config struct {
	Name    string
	Address string
	Options map[string]interface{}
	Timeout time.Duration
}

type Service struct {
	app.Service
	SendMessage *RemoteSendMessageTask
	Config      Config
	client      *kk.TCPClient
}

func (S *Service) Handle(a app.IApp, task app.ITask) error {
	return S.ReflectHandle(a, task, S)
}

func (S *Service) HandleInitTask(a app.IApp, task *app.InitTask) error {

	S.connect(a)

	return nil
}

func (S *Service) onMessage(a app.IApp, message *kk.Message) {

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
		app.Handle(a, &v)
		return
	}

	var name = message.To[len(S.client.Name()):]
	var tk, ok = app.NewTask(a, strings.Split(name, "."))

	log.Println(name)

	if !ok {
		var v = RemoteSendMessageTask{}
		v.Message = kk.Message{"NOIMPLEMENT", message.To, message.From, "text", []byte(name)}
		app.Handle(a, &v)
		return
	} else if message.Type == "text/json" || message.Type == "application/json" {
		var err = json.Decode(message.Content, tk)
		if err != nil {
			var b, _ = json.Encode(&app.Result{app.ERROR_UNKNOWN, "[json.Decode] [" + err.Error() + "] " + string(message.Content)})
			var v = RemoteSendMessageTask{}
			v.Message = kk.Message{message.Method, message.To, message.From, "text/json", b}
			app.Handle(a, &v)
			return
		}
	}

	go func() {
		var err = app.Handle(a, tk)
		if err != nil && err != app.Break {
			var b, _ = json.Encode(&app.Result{app.ERROR_UNKNOWN, err.Error()})
			var v = RemoteSendMessageTask{}
			v.Message = kk.Message{message.Method, message.To, message.From, "text/json", b}
			kk.GetDispatchMain().Async(func() {
				app.Handle(a, &v)
			})
			return
		} else {
			var b, _ = json.Encode(tk.GetResult())
			var v = RemoteSendMessageTask{}
			v.Message = kk.Message{message.Method, message.To, message.From, "text/json", b}
			kk.GetDispatchMain().Async(func() {
				app.Handle(a, &v)
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
		}, S.Config.Timeout*time.Second)
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
