package kk

import (
	"encoding/json"
	"log"
	"net"
	"strings"
	"time"
)

type TCPClient struct {
	Neuron
	chan_break   chan bool
	chan_message chan Message
	isconnected  bool

	OnMessage func(message *Message)

	OnConnected    func()
	OnDisconnected func(err error)
}

func (c *TCPClient) Break() {
	if c.chan_break != nil {
		c.chan_break <- true
	}
}

func (c *TCPClient) Send(message *Message, from INeuron) {
	if c.chan_message != nil {
		c.chan_message <- *message
	}
}

func (c *TCPClient) onDisconnected(err error) {
	if c.isconnected {
		c.isconnected = false
		if c.OnDisconnected != nil {
			c.OnDisconnected(err)
		}
	}
}

func (c *TCPClient) onConnected() {
	if !c.isconnected {
		c.isconnected = true
		if c.OnConnected != nil {
			c.OnConnected()
		}
	}
}

func NewTCPClient(name string, address string, options map[string]interface{}) *TCPClient {

	var v = TCPClient{}

	v.name = name
	v.address = address
	v.chan_message = make(chan Message)
	v.chan_break = make(chan bool)

	go func() {

		var conn, err = net.Dial("tcp", address)

		if err != nil {
			func(err error) {
				GetDispatchMain().Async(func() {
					if v.OnDisconnected != nil {
						v.OnDisconnected(err)
					}
				})
			}(err)
			close(v.chan_break)
			close(v.chan_message)
			return
		} else {
			GetDispatchMain().Async(func() {
				v.onConnected()
			})
		}

		var chan_rd = make(chan bool)
		var chan_wd = make(chan bool)

		go func() {

			var rd = NewMessageReader()

			for {

				var m, err = rd.Read(conn)

				if m != nil {
					func(message Message) {
						GetDispatchMain().Async(func() {

							if message.Method == "CONNECTED" {
								v.name = message.To
								log.Println("CONNECTED " + v.name)
							}

							if v.OnMessage != nil {
								v.OnMessage(&message)
							}
						})
					}(*m)
				} else if err != nil {
					func(err error) {
						GetDispatchMain().Async(func() {
							v.onDisconnected(err)
						})
					}(err)
					break
				}
			}

			select {
			case v.chan_break <- true:
			default:
			}

			chan_rd <- true

		}()

		go func() {

			var wd = NewMessageWriter()

			{
				var b []byte = nil
				if options != nil {
					b, _ = json.Marshal(options)
				}
				var m = Message{"CONNECT", name, "", "text/json", b}
				wd.Write(&m)
			}

			for {

				var r, err = wd.Done(conn)

				if err != nil {
					func(err error) {
						GetDispatchMain().Async(func() {
							v.onDisconnected(err)
						})
					}(err)
					break
				}

				if r {

					var m, ok = <-v.chan_message

					if !ok {
						break
					} else {
						wd.Write(&m)
					}

				}

			}

			select {
			case v.chan_break <- true:
			default:
			}

			chan_wd <- true
		}()

		<-v.chan_break

		close(v.chan_message)
		conn.Close()

		<-chan_rd
		<-chan_wd

		close(v.chan_break)
		close(chan_rd)
		close(chan_wd)

		v.chan_break = nil
		v.chan_message = nil
	}()

	return &v
}

func NewTCPClientConnection(conn net.Conn, id string) *TCPClient {

	var v = TCPClient{}

	v.name = ""
	v.address = conn.RemoteAddr().String()
	v.chan_message = make(chan Message)
	v.chan_break = make(chan bool)
	v.isconnected = true

	go func() {

		var chan_rd = make(chan bool)
		var chan_wd = make(chan bool)

		go func() {

			var rd = NewMessageReader()

			for {

				var m, err = rd.Read(conn)

				if m != nil {
					func(message Message) {
						GetDispatchMain().Async(func() {
							if message.Method == "CONNECT" {
								if strings.HasSuffix(message.From, ".*") {
									v.name = message.From[0:len(message.From)-1] + id
								} else {
									v.name = message.From
								}
								if message.Type == "text/json" && message.Content != nil {
									json.Unmarshal(message.Content, &v.options)
								}
								v.Send(&Message{"CONNECTED", v.name, v.name, "", []byte("")}, nil)
								log.Println("CONNECT " + v.name + " address: " + v.Address())
							} else if v.OnMessage != nil {
								v.OnMessage(&message)
							}
						})
					}(*m)
				} else if err != nil {
					func(err error) {
						GetDispatchMain().Async(func() {
							v.onDisconnected(err)
						})
					}(err)
					break
				}

			}

			select {
			case v.chan_break <- true:
			default:
			}

			chan_rd <- true

		}()

		go func() {

			var wd = NewMessageWriter()

			for {

				var r, err = wd.Done(conn)

				if err != nil {
					func(err error) {
						GetDispatchMain().Async(func() {
							v.onDisconnected(err)
						})
					}(err)
					break
				}

				if r {

					var m, ok = <-v.chan_message

					if !ok {
						break
					} else {
						wd.Write(&m)
					}

				}

			}

			select {
			case v.chan_break <- true:
			default:
			}

			chan_wd <- true
		}()

		<-v.chan_break

		close(v.chan_message)
		conn.Close()

		<-chan_rd
		<-chan_wd

		close(v.chan_break)
		close(chan_rd)
		close(chan_wd)

	}()

	return &v
}

func TCPClientConnect(name string, address string, options map[string]interface{}, onmessage func(message *Message)) (func(message *Message) bool, func() string) {

	var cli *TCPClient = nil
	var cli_connect func() = nil

	cli_connect = func() {

		log.Printf("Connect(%s) %s ...\n", name, address)

		cli = NewTCPClient(name, address, options)

		cli.OnConnected = func() {
			log.Printf("Connected(%s) %s \n", name, cli.Address())
		}

		cli.OnDisconnected = func(err error) {
			log.Printf("Disconnected(%s) %s %s\n", name, cli.Address(), err.Error())
			GetDispatchMain().AsyncDelay(cli_connect, time.Second)
		}

		cli.OnMessage = func(message *Message) {
			onmessage(message)
		}

	}

	cli_connect()

	return func(message *Message) bool {
			if cli != nil {
				cli.Send(message, nil)
				return true
			}
			return false
		}, func() string {
			if cli != nil {
				return cli.Name()
			}
			return name
		}

}
