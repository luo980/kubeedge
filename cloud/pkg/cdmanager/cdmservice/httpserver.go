package cdmservice

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/kubeedge/beehive/pkg/core/model"
	"github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/cdmcontext"
	"github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/cdmtype"
	"github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/common"
	"github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/coperate"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

var (
	//httpRequestMap		map[string]string
	httpRequestMapFull	map[string]map[string]string
)

type HttpServerWorker struct {
	Worker
	Group	string
}

var TransContext *cdmcontext.CDMContext

func (hsw HttpServerWorker) Start() {
	logrus.Error("Http ServerWorker Entered !!!")
	TransContext = hsw.CDMContexts

	initHttpRequestMap()
	logrus.Info("Map init Ok.")


	router := gin.Default()

	//Get Message realtime Query
	router.GET("/:name/:action", get)
	//Post Message for DB Operate
	router.POST("/:name/:action", post)

	err := router.Run(":80")
	if err != nil {
		logrus.Error("gin Server start failed.")
	}

	//initHttpRequestCallBack()
	//for {
	//	select {
	//	case msg, ok :=  <-hsw.ReceiverChan:
	//		if !ok {
	//		 }
	//	}
	//}

}

func get(gContext *gin.Context){
	name := gContext.Param("name")
	action := gContext.Param("action")
	fmt.Println( name, "+", action)
	logrus.Error( name ,"+", action)
	if _, exist := httpRequestMapFull[name][action]; !exist {
		gContext.Writer.WriteHeader(404)
		_, _ = gContext.Writer.Write([]byte("Can't recognize request."))
		return
	}
	switch name {
	case "edge":
		switch action {
		case httpRequestMapFull[common.EdgeModule][common.ShowEdges]:
			coperate.ShowEdges(gContext.Writer, gContext.Request)
		default:
		}
	case "device":
		switch action {
		case httpRequestMapFull[common.Device][common.ShowDevices]:
			coperate.ShowDevices(gContext.Writer, gContext.Request)
		case httpRequestMapFull[common.Device][common.QueryDevice]:
			coperate.QueryDevice(gContext.Writer, gContext.Request)
		default:
		}
	default:
		gContext.Writer.WriteHeader(404)
		_, _ = gContext.Writer.Write([]byte("Can't recognize request."))
	}
}

func post(gContext *gin.Context){
	name := gContext.Param("name")
	action := gContext.Param("action")
	if _, exist := httpRequestMapFull[name][action]; !exist {
		gContext.Writer.WriteHeader(404)
		_, _ = gContext.Writer.Write([]byte("Can't recognize request."))
		return
	}
	msg := CreatePostMessage(action, gContext)
	logrus.Info("Msg before send :", msg)
	TransContext.CommChan[name] <- msg
	rBack := "Receive " + name + "/" + action + " Request."
	gContext.Writer.WriteHeader(200)
	_, _ = gContext.Writer.Write([]byte(rBack))

}

//func initHttpRequestCallBack() {
//	httpRequestCallBack[common.AddDevice] 	=	AddDevice
//	httpRequestCallBack[common.AddEdge]		=	AddEdge
//	httpRequestCallBack[common.ShowEdges]	=	ShowEdges
//	httpRequestCallBack[common.ShowDevices] = 	ShowDevices
//	httpRequestCallBack[common.DeleteDevice] = 	DeleteDevice
//	httpRequestCallBack[common.QueryDevice]  =	QueryDevice
//}

func initHttpRequestMap(){
	//httpRequestMap = make(map[string]string)
	httpRequestMapFull = make(map[string]map[string]string)
	httpRequestMapFull[common.Edge] = make(map[string]string)
	httpRequestMapFull[common.Device] = make(map[string]string)
	initEdgeGetMap()
	initEdgePostMap()
	initDeviceGetMap()
	initDevicePostMap()
}

func initEdgeGetMap(){
	//httpRequestMap[common.ShowEdges]  =  "showEdges"
	httpRequestMapFull[common.Edge][common.ShowEdges] = common.ShowEdges
}

func initEdgePostMap(){
	//httpRequestMap[common.AddEdge]    =  "addEdge"
	httpRequestMapFull[common.Edge][common.AddEdge] = common.AddEdge
}

func initDeviceGetMap(){
	//httpRequestMap[common.ShowDevices] =   "showDevices"
	//httpRequestMap[common.QueryDevice]  =  "queryDevice"
	httpRequestMapFull[common.Device][common.ShowDevices] = common.ShowDevices
	httpRequestMapFull[common.Device][common.QueryDevice] = common.QueryDevice
}

func initDevicePostMap(){
	//httpRequestMap[common.AddDevice]   =  "addDevice"
	//httpRequestMap[common.DeleteDevice] =   "deleteDevice"
	httpRequestMapFull[common.Device][common.AddDevice] = common.AddDevice
	httpRequestMapFull[common.Device][common.DeleteDevice] = common.DeleteDevice
}

//func (hsw *HttpServerWorker) ShowEdges (gin *gin.Context){
//	Msg := model.NewMessage("")
//	Msg.Content = gin
//	hsMsg := cdmtype.CDMMessage{
//		Msg:      Msg,
//		Identity: "httpserver",
//		Action:   "ShowEdges",
//		Type:     "",
//	}
//	hsw.CDMContexts.CommChan[common.EdgeModule] <- hsMsg
//}

func CreateMessage (httpAct string, action string, gin *gin.Context) cdmtype.CDMMessage{
	Msg := model.NewMessage("")
	jsonRaw, err := ioutil.ReadAll(gin.Request.Body)
	if err != nil{
		logrus.Error("Read request body err:", err)
	}
	Msg.Content = jsonRaw
	hsMsg := cdmtype.CDMMessage{
		Msg:      Msg,
		Identity: "httpserver",
		Action:   action,
		Type:     httpAct,
	}
	return hsMsg
}

func CreatePostMessage (action string, gContext *gin.Context) cdmtype.ReqMMessage{
	content, _ := ioutil.ReadAll(gContext.Request.Body)
	hsMsg := cdmtype.ReqMMessage{
		Identity: "httpserver",
		Action:   action,
		Body: 	  content,
		RawQuery:	  gContext.Request.URL.RawQuery,
	}
	return hsMsg
}

//func showEdges(){
//
//}

//if  _, exist := httpRequestMap[action]; exist{
//	logrus.Info("Receive edge get request : ", action)
//	msg := CreateMessage(name, action, gContext)
//	TransContext.CommChan[common.EdgeModule] <- msg
//	logrus.Info("Send to EdgeChan")