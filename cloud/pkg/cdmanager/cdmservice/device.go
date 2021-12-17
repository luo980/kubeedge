package cdmservice

import (
	"github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/cdmtype"
	"github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/common"
	"github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/coperate"
	"github.com/sirupsen/logrus"
)

var (
	DeviceRequestAction map[string]string
)

type DevWorker struct {
	Worker
	Group string
}

func (dw DevWorker) Start(){
	initDeviceRequestAction()
	logrus.Error("Device Worker Entered !!!")
	for {
		select {
		case msg, ok := <-dw.ReceiverChan:
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
					"Query"		:	chanMsg.RawQuery,
				}).Infof("DevChan Msg content.")
				if _, exist := DeviceRequestAction[chanMsg.Action]; exist{
					switch chanMsg.Action {
					case common.AddDevice:
						err := coperate.AddDevice(chanMsg.Body)
						if err != nil{
							logrus.Error("Add Device Failed, err: ", err)
						}
					case common.DeleteDevice:
						err := coperate.DeleteDevice(chanMsg.RawQuery)
						if err != nil{
							logrus.Error("Delete Device Failed, err: ", err)
						}
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

func initDeviceRequestAction(){
	DeviceRequestAction = make(map[string]string)
	DeviceRequestAction[common.AddDevice] = common.AddDevice
	DeviceRequestAction[common.DeleteDevice] = common.DeleteDevice
}