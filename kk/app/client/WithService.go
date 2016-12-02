package client

import (
	"github.com/kkserver/kk-lib/kk/app"
	"time"
)

type WithService struct {
	app.Service
	Prefix     string
	Timeout    time.Duration
	InhertType string
}

func (S *WithService) Handle(a app.IApp, task app.ITask) error {

	v, ok := task.(IClientTask)

	if ok && v.GetInhertType() == S.InhertType {

		t := RequestTask{}

		t.Name = S.Prefix + v.GetClientName()
		t.Timeout = S.Timeout
		t.RequestTask = task

		return app.Handle(a, &t)

	}

	return nil
}
