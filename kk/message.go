package kk

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type Message struct {
	Method  string
	From    string
	To      string
	Type    string
	Content []byte
}

func (M *Message) String() string {
	if strings.HasPrefix(M.Type, "text") {
		return fmt.Sprintf("Method: %s From: %s To: %s Type: %s Content: %s", M.Method, M.From, M.To, M.Type, string(M.Content))
	}
	return fmt.Sprintf("Method: %s From: %s To: %s Type: %s Content: <%d>", M.Method, M.From, M.To, M.Type, len(M.Content))
}

type IReader interface {
	Read(data []byte) (int, error)
}

type IWriter interface {
	Write(data []byte) (int, error)
}

const MessageReaderStateKey = 0
const MessageReaderStateValue = 1
const MessageReaderStateContent = 2

type MessageReader struct {
	_state   int
	_key     *bytes.Buffer
	_value   *bytes.Buffer
	_length  int
	_content *bytes.Buffer
	_message Message
	_data    *bytes.Buffer
	_buf     []byte
}

func NewMessageReader() *MessageReader {
	var v = MessageReader{}
	v._state = MessageReaderStateKey
	v._key = bytes.NewBuffer(nil)
	v._value = bytes.NewBuffer(nil)
	v._length = 0
	v._content = bytes.NewBuffer(nil)
	v._data = bytes.NewBuffer(nil)
	v._buf = make([]byte, 20480)
	return &v
}

func (rd *MessageReader) readBytes() (*Message, error) {

	for rd._data.Len() > 0 {

		var c, err = rd._data.ReadByte()

		if err != nil {
			return nil, err
		}

		switch rd._state {
		case MessageReaderStateKey:
			{
				if c == ':' {
					rd._state = MessageReaderStateValue
					rd._value.Reset()
				} else if c == '\n' {
					if rd._length == 0 {
						rd._state = MessageReaderStateKey
						rd._message.Content = nil
						return &rd._message, nil
					} else {
						rd._state = MessageReaderStateContent
						rd._content.Reset()
					}
				} else {
					rd._key.WriteByte(c)
				}
			}
		case MessageReaderStateValue:
			{
				if c == '\n' {
					var key, _ = rd._key.ReadString(0)
					var value, _ = rd._value.ReadString(0)
					switch key {
					case "METHOD":
						rd._message.Method = value
					case "FROM":
						rd._message.From = value
					case "TO":
						rd._message.To = value
					case "TYPE":
						rd._message.Type = value
					case "LENGTH":
						rd._length, _ = strconv.Atoi(value)
					}
					rd._key.Reset()
					rd._value.Reset()
					rd._state = MessageReaderStateKey
				} else {
					rd._value.WriteByte(c)
				}
			}
		case MessageReaderStateContent:
			{
				rd._content.WriteByte(c)
				if rd._length == rd._content.Len() {
					rd._state = MessageReaderStateKey
					rd._message.Content = rd._content.Bytes()
					rd._length = 0
					rd._content.Reset()
					return &rd._message, nil
				}
			}
		}
	}

	return nil, nil
}

func (rd *MessageReader) Read(reader IReader) (*Message, error) {

	var v, err = rd.readBytes()

	if err != nil {
		return nil, err
	}

	if v == nil {

		var n, err = reader.Read(rd._buf)

		if err != nil {
			return nil, err
		}

		rd._data.Write(rd._buf[0:n])

		return rd.readBytes()
	}

	return v, err
}

type MessageWriter struct {
	_data *bytes.Buffer
}

func NewMessageWriter() *MessageWriter {
	var v = MessageWriter{}
	v._data = bytes.NewBuffer(nil)
	return &v
}

func (wd *MessageWriter) Done(writer IWriter) (bool, error) {

	if wd._data.Len() != 0 {
		var _, err = wd._data.WriteTo(writer)
		return false, err
	}

	return true, nil
}

func (wd *MessageWriter) Write(message *Message) {

	wd._data.WriteString("METHOD:")
	wd._data.WriteString(message.Method)
	wd._data.WriteByte('\n')

	wd._data.WriteString("FROM:")
	wd._data.WriteString(message.From)
	wd._data.WriteByte('\n')

	wd._data.WriteString("TO:")
	wd._data.WriteString(message.To)
	wd._data.WriteByte('\n')

	wd._data.WriteString("TYPE:")
	wd._data.WriteString(message.Type)
	wd._data.WriteByte('\n')

	wd._data.WriteString("LENGTH:")

	if message.Content == nil || len(message.Content) == 0 {
		wd._data.WriteString("0\n\n")
	} else {
		wd._data.WriteString(strconv.Itoa(len(message.Content)))
		wd._data.WriteString("\n\n")
		wd._data.Write(message.Content)
	}
}
