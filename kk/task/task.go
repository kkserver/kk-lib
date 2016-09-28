package task

type ITask interface {
}

type Task struct {
}

type AsyncTask struct {
	Done func()
	Fail func(err error)
}

type IAPITask interface {
	API() string
}
