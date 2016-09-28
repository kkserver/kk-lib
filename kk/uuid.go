package kk

import (
	"time"
)

const twepoch = int64(1424016000000)

var _id = twepoch

func milliseconds() int64 {
	return time.Now().UnixNano() / 1e6
}

func UUID() int64 {
	var id = milliseconds()
	for _id == id {
		id = milliseconds()
	}
	_id = id
	return _id - twepoch
}
