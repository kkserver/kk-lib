package kk

import (
	"fmt"
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
