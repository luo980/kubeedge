package dmservice

import (
	beehiveContext "github.com/kubeedge/beehive/pkg/core/context"
	"github.com/kubeedge/beehive/pkg/core/model"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmcommon"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmcontext"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmtype"
	"github.com/sirupsen/logrus"
)
var (
	//ActionCallBack map for action to callback
	ActionCallBack map[string]CallBack
)

type CommWorker struct {
	Worker
	Group string
}


func init() {
	initActionCallBack()
}

func (cw CommWorker) Start() {
	for{
		select {
		case msg, ok := <- cw.ReceiverChan:
			if !ok{
				return
			}
			if dmMsg, isDMMessage := msg.(*dmtype.DMMessage); isDMMessage{
				if fn, exist := ActionCallBack[dmMsg.Action]; exist{
					logrus.Info("Receive dmMsg: ", dmMsg, "Execute on :", dmMsg.Action)
					err := fn(cw.DMContexts, dmMsg.Identity, dmMsg.Msg)
					if err != nil {
						logrus.Error("Execute fn failed at Comm dmService, err: ", err)
					}
				}
			}
		}
	}
}


func initActionCallBack() {
	ActionCallBack = make(map[string]CallBack)
	ActionCallBack[dmcommon.SendToCloud] = dealSendToCloud
	ActionCallBack[dmcommon.SendToEdge] = dealSendToEdge
}

func dealSendToCloud(context *dmcontext.DMContext, resource string, msg interface{}) error {
	//if connect.CloudConnected == 0
	message, _ := msg.(*model.Message)
	beehiveContext.Send(dmcommon.HubModule, *message)
	return nil
}

func dealSendToEdge(context *dmcontext.DMContext, resource string, msg interface{}) error{
	return nil
}