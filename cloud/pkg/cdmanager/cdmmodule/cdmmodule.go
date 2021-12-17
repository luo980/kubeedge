package cdmmodule

import (
	"github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/cdmcontext"
	"github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/cdmservice"
	"github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/common"
	"github.com/sirupsen/logrus"
)

type CDMModule struct {
	Name string
	Worker cdmservice.CDMWorker
}

func (cdm *CDMModule) InitWorker(receive chan interface{}, confirm chan interface{}, cdmContext *cdmcontext.CDMContext){
	switch cdm.Name {
	case common.CommModule:
		cdm.Worker = cdmservice.CommWorker{
			Worker: cdmservice.Worker{
				ReceiverChan: receive,
				ConfirmChan:  confirm,
				CDMContexts:  cdmContext,
			},
			Group:  common.CommModule,
		}
	case common.DeviceModule:
		cdm.Worker = cdmservice.DevWorker{
			Worker: cdmservice.Worker{
				ReceiverChan: receive,
				ConfirmChan:  confirm,
				CDMContexts:  cdmContext,
			},
			Group:  common.DeviceModule,
		}
	case common.ServerModule:
		cdm.Worker = cdmservice.HttpServerWorker{
			Worker: cdmservice.Worker{
				ReceiverChan: receive,
				ConfirmChan:  confirm,
				CDMContexts:  cdmContext,
			},
			Group:  common.ServerModule,
		}
	case common.EdgeModule:
		cdm.Worker = cdmservice.EdgeWorker{
			Worker: cdmservice.Worker{
				ReceiverChan: receive,
				ConfirmChan:  confirm,
				CDMContexts:  cdmContext,
			},
			Group:  common.EdgeModule,
		}
	}
}

func (cdm CDMModule) Start(){
	defer func() {
		if err := recover(); err != nil{
			logrus.Info(cdm.Name, " in twin panic")
		}
	}()
	cdm.Worker.Start()
}