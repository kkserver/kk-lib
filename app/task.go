package app

type ITask interface {
}

type Task struct {
}

type AsyncTask struct {
	Done func(result interface{})
	Fail func(err error)
}

type IAPITask interface {
	API() string
	GetResult() interface{}
}

type Result struct {
	Errno  int    `json:"errno,omitempty"`
	Errmsg string `json:"errmsg,omitempty"`
}
