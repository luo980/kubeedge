package coperate

import (
	"context"
	"fmt"
	v1alpha2api "github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/cdeviceapi/client/v1alpha2"
	"github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/cdeviceapi/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"

	"xorm.io/xorm"
)

type CloudDeviceMeta struct {
	ID          	string `xorm:"pk"`
	DeviceName  	string
	DeviceModel 	string
	Manufacturer	string
	CreateTime		string
	//Description string
	Mac         	string
	Location    	string
	JoinTime		string
	EdgeName    	string
	Error       	string
}

type Model struct {
	ID           string `xorm:"pk"`
	Name         string
	Properties   string
	Description  string
	Type         string
	AccessMode   string
	DefaultValue string
}

func SyncDMetaFromAPIServer(e *xorm.Engine, crdClient *v1alpha2api.DevicesV1alpha2Client) error {
	e.DropTables("cloud_device_meta")
	errsyc := e.Sync2(new(CloudDeviceMeta))
	if errsyc != nil {
		print("Create table failed. err: ", errsyc)
	}

	DL := v1alpha2.DeviceList{}
	err := crdClient.RESTClient().Get().Resource("devices").Do(context.Background()).Into(&DL)

	if err != nil {
		return err
	}

	//i := v1alpha2.Device{}
	for key, item := range DL.Items {
		mac := ""
		location := ""
		joinTime := ""
		deviceID := ""
		manufacturer := ""
		//logrus.WithFields(logrus.Fields{
		// "\nindex"		: 	key,
		// "\nName"		:	item.Name,
		// "\nNamespace"	:	item.Namespace,
		// "\nDeviceModel":	item.Spec.DeviceModelRef.Name,
		// "\nNodeName"	:	item.Spec.NodeSelector.NodeSelectorTerms[0].MatchExpressions[0].Values[0],
		// "\nTopic"		:	item.Spec.Data.DataTopic,
		//}).Infof("\nHere's the device cached")
		print(
			"\nindex:\t\t", key,
			"\nAPI:\t\t", item.TypeMeta.APIVersion,
			"\nKind:\t\t", item.TypeMeta.Kind,
			"\nName:\t\t", item.Name,
			"\nNamespace:\t", item.Namespace,
			"\nDeviceModel:\t", item.Spec.DeviceModelRef.Name,
			"\nNodeName:\t", item.Spec.NodeSelector.NodeSelectorTerms[0].MatchExpressions[0].Values[0])

		for k, v := range item.ObjectMeta.Labels {
			if strings.Compare(k, "MAC") == 0 {
				mac = v
			}
			if strings.Compare(k, "JoinTime") == 0 {
				joinTime = v
			}
			if strings.Compare(k, "Location") == 0 {
				location = v
			}
			if strings.Compare(k, "DeviceID") == 0 {
				deviceID = v
			}
			if strings.Compare(k, "Manufacturer") == 0{
				manufacturer = v
			}
			print("\n", k, "\t", v)
		}
		for k, v := range item.Spec.Data.DataProperties {
			//logrus.WithFields(logrus.Fields{
			// "\nindex"		:	k,
			// "\nName"		: 	v.PropertyName,
			// "\nInfo"		:	v.Metadata,
			//}).Infof("\nHere's the device property")
			print(
				"\nindex\t", k,
				"\nName\t", v.PropertyName,
				"\nInfo\t", v.Metadata)
		}

		for k, v := range item.Status.Twins {
			//logrus.WithFields(logrus.Fields{
			// "\nindex"		:	k,
			// "  Name"		:	v.PropertyName,
			// "  type"		:	v.Reported.Metadata["0"],
			//}).Infof("\nHere's the device twins")
			print(
				"\n", k,
				"\t", v.PropertyName,
				"\t", v.Reported.Metadata["0"])
		}

		createTime := fmt.Sprint(item.CreationTimestamp)

		NewDevice := CloudDeviceMeta{
			ID:          	deviceID,
			DeviceName:     item.Name,
			DeviceModel:    item.Spec.DeviceModelRef.Name,
			Manufacturer: 	manufacturer,
			Mac: 			mac,
			JoinTime:       joinTime,
			CreateTime: 	createTime,
			Location:       location,
			EdgeName:       item.Spec.NodeSelector.NodeSelectorTerms[0].MatchExpressions[0].Values[0],
		}
		fmt.Println(NewDevice)
		result, err := e.Insert(NewDevice)
		if err != nil {
			return err
		}
		print("\nInsert Result is :", result)
	}
	return err
	//print("Request err is :", err.Error())
	//DeviceList, err := crdClient.Devices("default").List(context.Background(), v1.ListOptions{})
	//print("The return value is ", DeviceList)
	//print("The result is :", err)
	//if err != nil {
	// print("Crdclient get device list failed. err:", err)
	//}
	//for item := range DeviceList.Items {
	//print(item)
	//}
}

func SyncDeviceModelFromDB(e *xorm.Engine, crdClient *v1alpha2api.DevicesV1alpha2Client) error {
	errsyc := e.Sync2(new(Model))
	if errsyc != nil {
		print("Create table failed. err: ", errsyc)
	}

	DL := v1alpha2.DeviceModelList{}
	err := crdClient.RESTClient().Get().Resource("devicemodels").Do(context.Background()).Into(&DL)
	if err != nil {
		return err
	}

	for _, item := range DL.Items {
		newDeviceModel := Model{
			//ID:   "fuck",
			Name: item.ObjectMeta.Name,
		}
		for _, i := range item.Spec.Properties {
			newDeviceModel.Properties = i.Name
			newDeviceModel.Description = i.Description
			if i.Type.String != nil {
				newDeviceModel.Type = "string"
				newDeviceModel.AccessMode = string(i.Type.String.AccessMode)
				newDeviceModel.DefaultValue = i.Type.String.DefaultValue
			} else if i.Type.Boolean != nil {
				newDeviceModel.Type = "boolean"
				// if i.Type.Boolean.DefaultValue {
				// 	newDeviceModel.DefaultValue = string("true")
				// } else {
				// 	newDeviceModel.DefaultValue = string("false")
				// }
				newDeviceModel.DefaultValue = fmt.Sprint(i.Type.Boolean.DefaultValue)
			} else if i.Type.Bytes != nil {
				newDeviceModel.Type = "bytes"
				newDeviceModel.DefaultValue = ""
			} else if i.Type.Double != nil {
				newDeviceModel.Type = "double"
				newDeviceModel.AccessMode = string(i.Type.Double.AccessMode)
				newDeviceModel.DefaultValue = fmt.Sprint(i.Type.Double.DefaultValue)
			} else if i.Type.Float != nil {
				newDeviceModel.Type = "float"
			} else if i.Type.Int != nil {
				newDeviceModel.Type = "int"
				newDeviceModel.AccessMode = string(i.Type.Int.AccessMode)
				newDeviceModel.DefaultValue = fmt.Sprint(i.Type.Int.DefaultValue)

			}
			newDeviceModel.ID = time.Now().Format("2006-01-02 15:04:05.000000000")
			_, err := e.Insert(newDeviceModel)
			if err != nil {
				return err
			}
		}
	}
	return err
}

func DeleteDeviceFromDB(e *xorm.Engine, crdClient *v1alpha2api.DevicesV1alpha2Client, deviceID string) error{
	var deviceMeta CloudDeviceMeta
	ok, err := e.ID(deviceID).Get(&deviceMeta)
	if !ok {
		fmt.Println("Get deviceID err: ", err)
		return err
	}
	//result, err := crdClient.Devices("default").Create(context.Background(), &newDevice, metav1.CreateOptions{})
	err = crdClient.Devices("default").Delete(context.Background(), deviceMeta.DeviceName, metav1.DeleteOptions{})
	return err
}

func AddEdgeToDB(e *xorm.Engine, edge EdgeDB) error{
	err := e.Sync2(new(EdgeDB))
	if err != nil {
		print("Create table failed. err: ", err)
		return err
	}
	_, err = e.Insert(edge)
	return err
}

// func QueryDeviceFromDB(e *xorm.Engine, devicename string) {
// 	errsyc := e.Sync2(new(CloudDeviceMeta))
// 	if errsyc != nil {
// 		print("Create table failed. err: ", errsyc)
// 	}

// 	DL := v1alpha2.DeviceList{}
// 	err := crdClient.RESTClient().Get().Resource("devices").Do(context.Background()).Into(&DL)

// 	if err != nil {
// 		return err
// 	}
// }
