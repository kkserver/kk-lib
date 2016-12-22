package yaml

import (
	"reflect"
)

type Token interface {
}

type Decoder struct {
	data []byte
	idx  int
	B    struct {
	}
}

func NewDecoder(data []byte) {
	return &Decoder{data, 0}
}

func (D *Decoder) Next() (int, string, error) {

}
