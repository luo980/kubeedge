package cdmservice

import "github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/cdmcontext"

type CDMWorker interface {
	Start()
}

type Worker struct {
	ReceiverChan 	chan interface{}
	ConfirmChan		chan interface{}
	CDMContexts		*cdmcontext.CDMContext
}

type CallBack func(*cdmcontext.CDMContext, string, interface{}) error