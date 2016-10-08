package kkapp

import (
	"github.com/kkserver/kk-lib/app"
)

func New(parent app.IApp) *app.App {

	var v = app.NewApp(parent)

	v.Service(&KKService{})(&KKConnectTask{}, &KKDisconnectTask{}, &KKSendMessageTask{}, &KKReciveMessageTask{})

	return v
}
