package kk

type INeuron interface {
	Name() string
	Send(message *Message, from INeuron)
	Address() string
	Break()
	Get(key string) interface{}
	Set(key string, value interface{})
	Remove(key string)
}

type Neuron struct {
	name    string
	address string
	options map[string]interface{}
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
	var v = n.Get(key)
	if v == nil {
		return false
	}
	var b, ok = v.(bool)
	if ok {
		return b
	}
	return false
}

func (n *Neuron) GetInt(key string) int {
	var v = n.Get(key)
	if v == nil {
		return 0
	}
	var b, ok = v.(int)
	if ok {
		return b
	}
	return 0
}

func (n *Neuron) GetInt64(key string) int64 {
	var v = n.Get(key)
	if v == nil {
		return 0
	}
	var b, ok = v.(int64)
	if ok {
		return b
	}
	return 0
}

func (n *Neuron) GetString(key string) string {
	var v = n.Get(key)
	if v == nil {
		return ""
	}
	var b, ok = v.(string)
	if ok {
		return b
	}
	return ""
}

func (n *Neuron) Set(key string, value interface{}) {
	if n.options != nil {
		n.options = make(map[string]interface{})
	}
	n.options[key] = value
}

func (n *Neuron) Remove(key string) {
	if n.options != nil {
		delete(n.options, key)
	}
}
