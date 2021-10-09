package dmservice

import (
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmcontext"
)

//DMWorker worker for DManager
type DMWorker interface {
	Start()
}

//Worker actual
type Worker struct {
	ReceiverChan  chan interface{}
	ConfirmChan   chan interface{}
	HeartBeatChan chan interface{}
	DMContexts    *dmcontext.DMContext
}

//CallBack for deal
type CallBack func(*dmcontext.DMContext, string, interface{}) error
