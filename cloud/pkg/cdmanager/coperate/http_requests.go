package coperate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EdgeList struct {
	EdgeList map[string]Edge
}

type Edge struct {
	EdgeName string
	EdgeIP   string
	SSID     string
	Password string
}

type EdgeDB struct {
	EdgeName string `json:"EdgeName" xorm:"pk" xorm:"Text"`
	EdgeIp   string `json:"EdgeIp" xorm:"Text"`
	Ssid     string `json:"Ssid" xorm:"Text"`
	Password string `json:"Password" xorm:"Text"`
}
func (v *EdgeDB) TableName() string {
	return "EdgeList"
}

type RegDevice struct {
	DeviceType  	string
	DeviceName  	string
	DeviceModel 	string
	Manufacturer	string
	DeviceID    	string
	MAC         	string
	Location    	string
	JoinTime    	string
	EdgeName    	string
}



var e, _ = EngineInit()
var crdClient, _ = InitCrdClient()

func AddDevice(content []byte) error{
	var recMsg RegDevice
	err := json.Unmarshal(content, &recMsg)
	if err != nil {
		fmt.Println("Unmarshal Error :", err)
	}

	newDevice := CreateNewDevice(recMsg)
	result, err := crdClient.Devices("default").Create(context.Background(), &newDevice, metav1.CreateOptions{})
	//result, err := newDeviceClient.Create(context.Background(), &newDevice, v1.CreateOptions{})
	if err != nil {
		fmt.Println("Create failed, err :", err)
		return err
	}
	fmt.Println("Create result :", result)
	err = SyncDMetaFromAPIServer(e, crdClient)
	return err
}

func DeleteDevice(rawQuery string) error {
	//result := make(map[string]string)
	keys, _ := url.ParseQuery(rawQuery)
	for k, v := range keys {
		switch k {
		case "DeviceID":
			fmt.Println("Receive a delete Request, DeviceID: ", v)
			for _, v1 := range v{
				err := DeleteDeviceFromDB(e, crdClient, v1)
				if err != nil {
					return err
				}
				err = SyncDMetaFromAPIServer(e, crdClient)
				return err
			}
		}
	}
	return nil
}

func Get(w http.ResponseWriter, req *http.Request) {
	var message []byte
	//result := make(map[string]string)
	keys := req.URL.Query()
	for _, v := range keys {
		switch v[0] {
		case "getEdgeList":
			fmt.Println("Receive a list request")
			sql := "select * from EdgeList"
			EdgeList := GetQListFromDB(sql, e)
			message, _ = json.Marshal(EdgeList)
			w.WriteHeader(200)
			w.Header().Set("Content-Type", "application/json")
		}
	}
	//log.Println(result)
	w.Write(message)
}

func ShowEdges(w http.ResponseWriter, req *http.Request) {
	sql := "select * from EdgeList"
	EdgeList := GetQListFromDB(sql, e)
	message, _ := json.Marshal(EdgeList)
	w.Header().Set("Content-Type", "application/json")
	w.Write(message)
}

func Post(w http.ResponseWriter, req *http.Request) {
	result := make(map[string]string)
	keys := req.URL.Query()
	for k, v := range keys {
		result[k] = v[0]
	}
	log.Println(result)
}

func ShowDevices(w http.ResponseWriter, req *http.Request){
	sql := "select * from cloud_device_meta"
	DeviceList := GetQListFromDB(sql, e)
	message, _ := json.Marshal(DeviceList)
	w.Header().Set("Content-Type", "application/json")
	w.Write(message)
	//if err != nil{
	//	w.Write([]byte(err.Error()))
	//}
	//fmt.Println("sql result is : ", results)
	//message, _ := json.Marshal(results)
	////result := fmt.Sprint(results)
	//w.WriteHeader(200)
	//w.Write([]byte(message))
}

func QueryDevice(w http.ResponseWriter, req *http.Request) {
	var message []byte
	//result := make(map[string]string)
	keys := req.URL.Query()
	for k, v := range keys {
		switch k {
		case "DeviceID":
			fmt.Println("Receive a delete Request, DeviceID: ", v)
			for _, v1 := range v {
				var deviceMeta CloudDeviceMeta
				ok, err := e.ID(v1).Get(&deviceMeta)
				if !ok {
					w.Write([]byte(err.Error()))
				}
				message, _ = json.Marshal(deviceMeta)
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(message)
}

func AddEdge(content []byte) error{
	var recMsg EdgeDB
	err := json.Unmarshal(content, &recMsg)
	if err != nil {
		fmt.Println("Unmarshal Error :", err)
	}
	err = AddEdgeToDB(e, recMsg)
	return err
}

func HelloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "helllo, world.\n")
	fmt.Println("have one hello")
}

func HelloServer2(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "helllo, world 2.\n")
	fmt.Println("have one hello")
}
