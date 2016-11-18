package json

import (
	"encoding/json"
)

func Encode(object interface{}) ([]byte, error) {
	return json.Marshal(object)
}
