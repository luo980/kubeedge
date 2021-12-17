package cdmservice

import (
	"github.com/gin-gonic/gin"
	beehiveContext "github.com/kubeedge/beehive/pkg/core/context"
	"github.com/kubeedge/beehive/pkg/core/model"
	"github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/cdmcontext"
	"github.com/sirupsen/logrus"
)

var (
	ActionCallBack map[string]CallBack
)

type CommWorker struct {
	Worker
	Group string
}

func init(){
	initActionCallBack()
}

func (cw CommWorker) Start()  {
	logrus.Error("Comm Worker Start !!!")
}

func dealSendToEdge(gContext *gin.Context, context *cdmcontext.CDMContext, resource string, msg interface{}) error {
	sendMsg, ok := msg.(*model.Message)
	if !ok {
		logrus.Error("Not a model.message msg")
	}
	beehiveContext.Send("cloudhub" ,*sendMsg)
	return nil
}


func initActionCallBack(){
	ActionCallBack = make(map[string]CallBack)
	//ActionCallBack[common.SendToEdge] = dealSendToCloud
	//ActionCallBack[common.SendToEdge] = dealSendToEdge
}