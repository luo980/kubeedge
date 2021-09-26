package dmmodule

import (
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmcontext"
	"k8s.io/klog/v2"
)

//DMModule module for DManager
type DMModule struct {
	Name   string
	Worker DMWorker
}

//testWorker
type testWorker struct {
	Worker
	Group string
}

//InitWorker init workers
func (dm *DMModule) InitWorker(recv chan interface{}, confirm chan interface{}, heartBeat chan interface{}, dmContext *dmcontext.DMContext) {
	dm.Worker = testWorker{
		Group: "test",
		Worker: Worker{
			ReceiverChan:  recv,
			ConfirmChan:   confirm,
			HeartBeatChan: heartBeat,
			DMContexts:    dmContext,
		},
	}
}

func (tw testWorker) Start() {
	// o something
	klog.Infof("DIY DM Worker start!")
}

//Start module, actual worker start
func (dm DMModule) Start() {
	defer func() {
		if err := recover(); err != nil {
			klog.Infof("%s in twin panic", dm.Name)
		}
	}()
	dm.Worker.Start()
}
