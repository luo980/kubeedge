package cdmservice

import (
	"github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/cdmtype"
	"github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/common"
	"github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/coperate"
	"github.com/sirupsen/logrus"
)

//var (
//	EdgeActionCallBack map[string]CallBack
//)

type EdgeWorker struct {
	Worker
	Group string
}

var (
	EdgeRequestAction map[string]string
)

func (ew EdgeWorker) Start(){
	initEdgeRequestAction()
	logrus.Error("Edge Worker Entered !!!")
	for {
		select {
		case msg, ok := <-ew.ReceiverChan:
			if !ok {
				return
			}
			logrus.Error("Msg: ", msg)
			a, b := msg.(cdmtype.ReqMMessage)
			logrus.Error("a: ", a, ". b: ", b)
			if chanMsg, isCDMMessage := msg.(cdmtype.ReqMMessage); isCDMMessage{
				logrus.Info("Receive Msg from Chan !")
				logrus.WithFields(logrus.Fields{
					"Identity"  : 	chanMsg.Identity,
					"Action"	:	chanMsg.Action,
					"Type"		:	chanMsg.Type,
					"Body"		:	string(chanMsg.Body),
				}).Infof("EdgeChan Msg content.")
				if _, exist := EdgeRequestAction[chanMsg.Action]; exist{
					switch chanMsg.Action {
					case common.AddEdge:
						err := coperate.AddEdge(chanMsg.Body)
						if err != nil {
							logrus.Error("AddEdge Failed: ", err)
						}
					default:
					}
				}
				//logrus.Info("Body is : ", string(chanMsg.Body))

				//if ginMsg, isGinMsg := chanMsg.Msg.Content.(*gin.Context); isGinMsg{
				//	jsonData, err := ioutil.ReadAll(ginMsg.Request.Body)
				//	if err != nil {
				//		logrus.Error("Err in reading request body.")
				//	}
				//	//err = json.Unmarshal()
				//	logrus.Info("GinMsg content: ", jsonData)
				//}
			}
		}
	}

}



//func initEdgeActionCallBack(){
//	//EdgeActionCallBack = make(map[string]CallBack)
//	//ActionCallBack[common.SendToEdge] = dealSendToCloud
//	//ActionCallBack[common.SendToEdge] = dealSendToEdge
//}

func initEdgeRequestAction(){
	EdgeRequestAction = make(map[string]string)
	EdgeRequestAction[common.AddEdge] = common.AddEdge
}