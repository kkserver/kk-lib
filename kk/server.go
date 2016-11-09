package kk

import (
	"container/list"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

type TCPServer struct {
	Neuron
	chan_break     chan bool
	clients        *list.List
	OnMessage      func(message *Message, from INeuron)
	OnStart        func()
	OnFail         func(err error)
	OnAccept       func(client *TCPClient)
	OnDisconnected func(client *TCPClient, err error)
}

func (c *TCPServer) Break() {
	c.chan_break <- true
}

func (c *TCPServer) Send(message *Message, from INeuron) {

	if message.Method == "SERVER" && message.To == c.name {
		if from != nil {
			var m = Message{"CLIENT", "", message.From, "text", []byte("")}
			var e = c.clients.Front()
			for e != nil {
				var f = e.Value.(*TCPClient)
				m.From = f.name
				m.Content = []byte(f.address)
				if m.From != "" {
					from.Send(&m, nil)
				}
				e = e.Next()
			}
		}
		return
	}

	var e = c.clients.Back()
	var count = 0
	for e != nil {
		var f = e.Value.(*TCPClient)
		if (from == nil || (from != f)) && f.name != "" && strings.HasPrefix(message.To, f.name) {
			f.Send(message, from)
			count += 1
			if f.GetBoolean("exclusive") {
				break
			}
		}
		e = e.Prev()
	}

	if count == 0 && from != nil && c.name != message.To {
		var m = Message{"DONE", c.name, message.From, "text", []byte(message.Method)}
		from.Send(&m, nil)
	}

}

func NewTCPServer(name string, address string, maxconnections int) *TCPServer {

	var v = TCPServer{}

	v.name = name
	v.address = address
	v.chan_break = make(chan bool, 8)
	v.clients = list.New()

	var uuid int64 = time.Now().UnixNano()

	go func() {

		var listen, err = net.Listen("tcp", address)

		if err != nil {
			func(err error) {
				GetDispatchMain().Async(func() {
					if v.OnFail != nil {
						v.OnFail(err)
					}
				})
			}(err)
			return
		} else {
			GetDispatchMain().Async(func() {
				if v.OnStart != nil {
					v.OnStart()
				}
			})
		}

		var num_connections = 0
		var chan_num_connections = make(chan bool, 2048)

		var chan_accept = make(chan bool, 2048)

		go func() {

			for {

				for num_connections >= maxconnections {
					log.Printf("wait %d:%d ...\n", num_connections, maxconnections)
					var v, ok = <-chan_num_connections
					if !v || !ok {
						return
					}
				}

				log.Println("accept ...")

				var conn, err = listen.Accept()

				if err != nil {
					func(err error) {
						GetDispatchMain().Async(func() {
							if v.OnFail != nil {
								v.OnFail(err)
							}
						})
					}(err)
					return
				}

				if conn == nil {
					return
				}

				func(conn net.Conn) {

					GetDispatchMain().Async(func() {

						num_connections += 1

						log.Printf("connections: %d\n", num_connections)

						uuid = uuid + 1

						var client = NewTCPClientConnection(conn, strconv.FormatInt(uuid, 10))

						v.clients.PushBack(client)

						client.OnDisconnected = func(err error) {
							if v.OnDisconnected != nil {
								v.OnDisconnected(client, err)
							}
							var e = v.clients.Front()
							for e != nil {
								var f = e.Value.(*TCPClient)
								if f == client {
									var n = e.Next()
									v.clients.Remove(e)
									e = n
									break
								}
								e = e.Next()
							}
							num_connections -= 1
							select {
							case chan_num_connections <- true:
							default:
							}
							log.Printf("connections: %d\n", num_connections)
						}
						client.OnMessage = func(message *Message) {
							if v.OnMessage != nil {
								v.OnMessage(message, client)
							}
						}
						if v.OnAccept != nil {
							v.OnAccept(client)
						}
					})
				}(conn)
			}

			select {
			case v.chan_break <- true:
			default:
			}
		}()

		<-v.chan_break

		listen.Close()
		close(chan_num_connections)

		<-chan_accept

		close(v.chan_break)

		GetDispatchMain().Async(func() {
			var e = v.clients.Front()
			for e != nil {
				var f = e.Value.(*TCPClient)
				f.Break()
				e = e.Next()
			}
		})

	}()

	return &v
}
