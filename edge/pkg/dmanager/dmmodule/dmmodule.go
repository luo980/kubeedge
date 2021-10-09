package dmmodule

import (
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmcommon"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmcontext"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmservice"
	"k8s.io/klog/v2"
)

//DMModule module for DManager
type DMModule struct {
	Name   string
	Worker dmservice.DMWorker
}

//testWorker
type testWorker struct {
	dmservice.Worker
	Group string
}

//InitWorker init workers
func (dm *DMModule) InitWorker(recv chan interface{}, confirm chan interface{}, heartBeat chan interface{}, dmContext *dmcontext.DMContext) {
	switch dm.Name {
	case dmcommon.TestModule:
		dm.Worker = testWorker{
			Group: "test",
			Worker: dmservice.Worker{
				ReceiverChan:  recv,
				ConfirmChan:   confirm,
				HeartBeatChan: heartBeat,
				DMContexts:    dmContext,
			},
		}
	case dmcommon.DeviceModule:
		dm.Worker = dmservice.DeviceWorker{
			Group: dmcommon.DeviceModule,
			Worker: dmservice.Worker{
				ReceiverChan:  recv,
				ConfirmChan:   confirm,
				HeartBeatChan: heartBeat,
				DMContexts:    dmContext,
			},
		}
	case dmcommon.CommModule:
		dm.Worker = dmservice.CommWorker{
			Group: dmcommon.CommModule,
			Worker: dmservice.Worker{
				ReceiverChan:  recv,
				ConfirmChan:   confirm,
				HeartBeatChan: heartBeat,
				DMContexts:    dmContext,
			},
		}
	case dmcommon.MemModule:
		dm.Worker = dmservice.MemWorker{
			Group: dmcommon.MemModule,
			Worker: dmservice.Worker{
				ReceiverChan:  recv,
				ConfirmChan:   confirm,
				HeartBeatChan: heartBeat,
				DMContexts:    dmContext,
			},
		}
	case dmcommon.AbilityModule:
		dm.Worker = dmservice.AbiWorker{
			Group: dmcommon.AbilityModule,
			Worker: dmservice.Worker{
				ReceiverChan:  recv,
				ConfirmChan:   confirm,
				HeartBeatChan: heartBeat,
				DMContexts:    dmContext,
			},
		}
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
