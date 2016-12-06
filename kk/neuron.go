package kk

import (
	Value "github.com/kkserver/kk-lib/kk/value"
	"reflect"
)

type INeuron interface {
	Name() string
	Send(message *Message, from INeuron)
	Address() string
	Break()
	Get(key string) interface{}
	Set(key string, value interface{})
	Remove(key string)
	Options() map[string]interface{}
}

type Neuron struct {
	name    string
	address string
	options map[string]interface{}
}

func (n *Neuron) Options() map[string]interface{} {
	if n.options == nil {
		n.options = map[string]interface{}{}
	}
	return n.options
}

func (n *Neuron) Name() string {
	return n.name
}

func (n *Neuron) Address() string {
	return n.address
}

func (n *Neuron) Get(key string) interface{} {
	if n.options != nil {
		return n.options[key]
	}
	return nil
}

func (n *Neuron) GetBoolean(key string) bool {
	return Value.BooleanValue(Value.Get(reflect.ValueOf(n.options), key), false)
}

func (n *Neuron) GetInt(key string) int64 {
	return Value.IntValue(Value.Get(reflect.ValueOf(n.options), key), 0)
}

func (n *Neuron) GetString(key string) string {
	return Value.StringValue(Value.Get(reflect.ValueOf(n.options), key), "")
}

func (n *Neuron) Set(key string, value interface{}) {
	if n.options == nil {
		n.options = map[string]interface{}{}
	}
	n.options[key] = value
}

func (n *Neuron) Remove(key string) {
	if n.options != nil {
		delete(n.options, key)
	}
}
