package kk

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"github.com/kkserver/kk-lib/kk/json"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
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
	if c.isconnected && c.chan_message != nil {
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
	v.chan_message = make(chan Message, 2048)
	v.chan_break = make(chan bool, 8)

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

		var chan_rd = make(chan bool, 8)
		var chan_wd = make(chan bool, 8)

		go func() {

			var rd = bufio.NewReader(conn)
			var dec = gob.NewDecoder(rd)

			for {

				var message = Message{}

				var err = dec.Decode(&message)

				if err != nil {
					func(err error) {
						GetDispatchMain().Async(func() {
							v.onDisconnected(err)
						})
					}(err)
					break
				}

				func(message Message) {

					GetDispatchMain().Async(func() {

						if message.Method == "CONNECTED" {
							v.name = message.To
							log.Println("CONNECTED " + v.name + " " + string(message.Content))
						}

						if v.OnMessage != nil {
							v.OnMessage(&message)
						}
					})

				}(message)
			}

			select {
			case v.chan_break <- true:
			default:
			}

			chan_rd <- true

		}()

		go func() {

			var wd = bufio.NewWriter(conn)
			var enc = gob.NewEncoder(wd)

			{
				var b []byte = nil
				if options != nil {
					b, _ = json.Encode(options)
				}
				var m = Message{"CONNECT", name, "", "text/json", b}
				enc.Encode(m)
			}

			for {

				var err = wd.Flush()

				if err != nil {
					func(err error) {
						GetDispatchMain().Async(func() {
							v.onDisconnected(err)
						})
					}(err)
					break
				}

				var message, ok = <-v.chan_message

				if !ok {
					break
				} else {
					enc.Encode(message)
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
	v.chan_message = make(chan Message, 2048)
	v.chan_break = make(chan bool, 8)
	v.isconnected = true

	go func() {

		var chan_rd = make(chan bool, 8)
		var chan_wd = make(chan bool, 8)

		go func() {

			var rd = bufio.NewReader(conn)
			var dec = gob.NewDecoder(rd)

			for {

				var message = Message{}

				var err = dec.Decode(&message)

				if err != nil {
					func(err error) {
						GetDispatchMain().Async(func() {
							v.onDisconnected(err)
						})
					}(err)
					break
				} else {
					func(message Message) {
						GetDispatchMain().Async(func() {
							if message.Method == "CONNECT" {
								if strings.HasSuffix(message.From, ".*") {
									v.name = message.From[0:len(message.From)-1] + id
								} else {
									v.name = message.From
								}
								if message.Type == "text/json" && message.Content != nil {
									json.Decode(message.Content, &v.options)
								}
								v.Send(&Message{"CONNECTED", v.name, v.name, "text", []byte(v.Address())}, nil)
								log.Println("CONNECT " + v.name + " address: " + v.Address())
								if v.OnConnected != nil {
									v.OnConnected()
								}
							} else if v.OnMessage != nil {
								v.OnMessage(&message)
							}
						})
					}(message)
				}

			}

			select {
			case v.chan_break <- true:
			default:
			}

			chan_rd <- true

		}()

		go func() {

			var wd = bufio.NewWriter(conn)
			var enc = gob.NewEncoder(wd)

			for {

				var message, ok = <-v.chan_message

				if !ok {
					break
				} else {
					enc.Encode(message)
				}

				var err = wd.Flush()

				if err != nil {
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
	var isConnected = make(chan bool, 8)

	cli_connect = func() {

		log.Printf("Connect(%s) %s ...\n", name, address)

		cli = NewTCPClient(name, address, options)

		cli.OnConnected = func() {
			log.Printf("Connected(%s) %s \n", name, cli.Address())
			select {
			case isConnected <- true:
			default:
			}
		}

		cli.OnDisconnected = func(err error) {
			log.Printf("Disconnected(%s) %s %s\n", name, cli.Address(), err.Error())
			cli = nil
			GetDispatchMain().AsyncDelay(cli_connect, time.Second)
		}

		cli.OnMessage = func(message *Message) {
			onmessage(message)
		}

	}

	cli_connect()

	return func(message *Message) bool {
			if cli == nil {
				<-isConnected
			}
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

func TCPClientRequestConnect(name string, address string, options map[string]interface{}) (func(message *Message, trackId string, timeout time.Duration) *Message, func() string, func(message *Message)) {

	var https = map[int64]chan Message{}

	var sendMessage, getName = TCPClientConnect(name, address, options, func(message *Message) {

		var i = strings.LastIndex(message.To, ".")
		var id, _ = strconv.ParseInt(message.To[i+1:], 10, 64)
		var ch, ok = https[id]

		if ok && ch != nil {
			if message.Method == "REQUEST" {
				ch <- *message
				delete(https, id)
			} else {
				var m = Message{"UNAVAILABLE", "", "", "", []byte("")}
				ch <- m
				delete(https, id)
			}
		}
	})

	var uuid int64 = time.Now().UnixNano()

	return func(message *Message, trackId string, timeout time.Duration) *Message {

			var id int64 = 0
			var ch = make(chan Message, 2048)
			defer close(ch)

			GetDispatchMain().Async(func() {

				id = uuid + 1
				uuid = id
				https[id] = ch

				message.From = fmt.Sprintf("%s%s.%d", getName(), trackId, id)
				message.Method = "REQUEST"

				if !sendMessage(message) {
					var r = Message{"TIMEOUT", "", "", "", []byte("")}
					ch <- r
					delete(https, id)
				}

			})

			GetDispatchMain().AsyncDelay(func() {

				var ch = https[id]

				if ch != nil {
					var r = Message{"TIMEOUT", "", "", "", []byte("")}
					ch <- r
					delete(https, id)
				}

			}, timeout)

			var m, ok = <-ch

			if !ok {
				var r = Message{"TIMEOUT", "", "", "", []byte("")}
				return &r
			} else {
				return &m
			}

		}, getName, func(message *Message) {

			GetDispatchMain().Async(func() {

				if message.From == "" {
					message.From = getName()
				}

				sendMessage(message)

			})

		}
}

func TCPClientHandleFunc(name string, address string, options map[string]interface{}, alias string, timeout time.Duration) func(w http.ResponseWriter, r *http.Request) {

	request, getname, sendMessage := TCPClientRequestConnect(name, address, options)

	return func(w http.ResponseWriter, r *http.Request) {

		var body = make([]byte, r.ContentLength)
		var contentType = r.Header.Get("Content-Type")
		var to = r.RequestURI[len(alias):]
		var n, err = r.Body.Read(body)
		defer r.Body.Close()

		if err != nil && err != io.EOF {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		} else if int64(n) != r.ContentLength {
			log.Printf("%d %d\n", n, r.ContentLength)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var trackId = ""

		{
			var ip = r.Header.Get("X-CLIENT-IP")

			if ip == "" {
				ip = r.Header.Get("X-Real-IP")
			}

			if ip == "" {
				ip = r.RemoteAddr
			}

			var cookie, err = r.Cookie("kk")

			if err != nil {
				var v = http.Cookie{}
				v.Name = "kk"
				v.Value = strconv.FormatInt(time.Now().UnixNano(), 10)
				v.Expires = time.Now().Add(24 * 3600 * time.Second)
				v.HttpOnly = true
				v.MaxAge = 24 * 3600
				v.Path = "/"
				http.SetCookie(w, &v)
				cookie = &v
			}

			trackId = cookie.Value

			var b, _ = json.Encode(map[string]string{"code": trackId, "ip": ip,
				"User-Agent": r.Header.Get("User-Agent"),
				"Referer":    r.Header.Get("Referer"),
				"Path":       r.RequestURI,
				"Host":       r.Host,
				"Protocol":   r.Proto})

			var m = Message{"MESSAGE", getname(), "kk.message.http.request", "text/json", b}

			GetDispatchMain().Async(func() {
				sendMessage(&m)
			})

		}

		go func() {

			var r = request(&Message{"REQUEST", "", to, contentType, body}, trackId, timeout)

			if r.Method == "TIMEOUT" {
				w.WriteHeader(http.StatusGatewayTimeout)
			} else if r.Method == "UNAVAILABLE" {
				w.WriteHeader(http.StatusServiceUnavailable)
			} else if r.Method == "REQUEST" {
				w.Header().Add("From", r.From)
				if strings.HasPrefix(r.Type, "text") {
					w.Header().Add("Content-Type", r.Type+"; charset=utf-8")
				} else {
					w.Header().Add("Content-Type", r.Type)
				}
				w.Header().Add("Content-Length", strconv.Itoa(len(r.Content)))
				w.WriteHeader(http.StatusOK)
				w.Write(r.Content)
			} else {
				w.WriteHeader(http.StatusUnsupportedMediaType)
			}

		}()

	}

}
