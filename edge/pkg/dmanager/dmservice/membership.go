package dmservice

import (
	"errors"
	"fmt"
	"github.com/kubeedge/beehive/pkg/core/model"
	messagepkg "github.com/kubeedge/kubeedge/edge/pkg/common/message"
	"github.com/kubeedge/kubeedge/edge/pkg/common/modules"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmcommon"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmcontext"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmdatabase"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmtype"
	"github.com/sirupsen/logrus"
	"k8s.io/klog/v2"
	"strings"
	"sync"
	"time"
)

var (
	//memActionCallBack map for action to callback
	memActionCallBack map[string]CallBack
)

type MemWorker struct {
	Worker
	Group string
}

func (mw MemWorker) Start() {
	initMemActionCallBack()
	for {
		select {
		case msg, ok := <-mw.ReceiverChan:
			if !ok {
				return
			}
			if dmMsg, isDMMessage := msg.(*dmtype.DMMessage); isDMMessage {
				if fn, exist := memActionCallBack[dmMsg.Action]; exist {
					logrus.WithFields(logrus.Fields{
						"module":   "dmservice",
						"func":     "Memworker Start()",
						"Identity": dmMsg.Identity,
						"Msg":      dmMsg.Msg,
						"Action":   dmMsg.Action,
					}).Infof("membership memActionCallBack isExist?")
					err := fn(mw.DMContexts, dmMsg.Identity, dmMsg.Msg)
					if err != nil {
						klog.Errorf("MemModule deal %s event failed: %v", dmMsg.Action, err)
					}
				} else {
					klog.Errorf("MemModule deal %s event failed, not found callback", dmMsg.Action)
				}
			}

		case v, ok := <-mw.HeartBeatChan:
			if !ok {
				return
			}
			if err := mw.DMContexts.HeartBeat(mw.Group, v); err != nil {
				return
			}
		}
	}
}

func initMemActionCallBack() {
	memActionCallBack = make(map[string]CallBack)
	memActionCallBack[dmcommon.MemGet] = dealMembershipGet
	memActionCallBack[dmcommon.MemUpdated] = dealMembershipUpdate
	memActionCallBack[dmcommon.MemDetailResult] = dealMembershipDetail

}

func dealMembershipGet(context *dmcontext.DMContext, resource string, msg interface{}) error {
	klog.Infof("MEMBERSHIP EVENT")
	message, ok := msg.(*model.Message)
	if !ok {
		return errors.New("msg not Message type")
	}

	contentData, ok := message.Content.([]byte)
	if !ok {
		return errors.New("assertion failed")
	}

	dealMembershipGetInner(context, contentData)
	return nil
}

func dealMembershipUpdate(context *dmcontext.DMContext, resource string, msg interface{}) error {
	klog.Infof("Membership event")
	message, ok := msg.(*model.Message)
	if !ok {
		return errors.New("msg not Message type")
	}

	contentData, ok := message.Content.([]byte)
	if !ok {
		return errors.New("assertion failed")
	}

	logrus.WithFields(logrus.Fields{
		"module":  "dmservice",
		"func":    "dealMembershipUpdate()",
		"message": message,
		//"contentData" : contentData,
	}).Infof("MemUpdated actions")

	updateEdgeGroups, err := dmtype.UnmarshalMembershipUpdate(contentData)
	logrus.WithFields(logrus.Fields{
		"module": "dmservice",
		"func":   "UnmarshalMembershipUpdate()",
		"update": updateEdgeGroups,
	}).Infof("UnmarshalMembershipUpdate actions")

	if err != nil {
		klog.Errorf("Unmarshal membership info failed , err: %#v", err)
		return err
	}

	baseMessage := dmtype.BaseMessage{EventID: updateEdgeGroups.EventID}
	logrus.WithFields(logrus.Fields{
		"module":                         "dmservice",
		"baseMessage":                    baseMessage,
		"updateEdgeGroups.EventID":       updateEdgeGroups.EventID,
		"updateEdgeGroups.AddDevices":    updateEdgeGroups.AddDevices,
		"updateEdgeGroups.RemoveDevices": updateEdgeGroups.RemoveDevices,
	}).Infof("EventID and Add/Remove")
	if updateEdgeGroups.AddDevices != nil && len(updateEdgeGroups.AddDevices) > 0 {
		//add device
		addDevice(context, updateEdgeGroups.AddDevices, baseMessage, false)

	}
	if updateEdgeGroups.RemoveDevices != nil && len(updateEdgeGroups.RemoveDevices) > 0 {
		// delete device
		removeDevice(context, updateEdgeGroups.RemoveDevices, baseMessage, false)
	}
	return nil
}

func dealMembershipDetail(context *dmcontext.DMContext, resource string, msg interface{}) error {
	klog.Info("Deal node detail info")
	message, ok := msg.(*model.Message)
	if !ok {
		return errors.New("msg not Message type")
	}

	contentData, ok := message.Content.([]byte)
	if !ok {
		return errors.New("assertion failed")
	}

	devices, err := dmtype.UnmarshalMembershipDetail(contentData)
	if err != nil {
		klog.Errorf("Unmarshal membership info failed , err: %#v", err)
		return err
	}

	baseMessage := dmtype.BaseMessage{EventID: devices.EventID}
	defer context.UnlockAll()
	context.LockAll()
	var toRemove []dmtype.Device
	isDelta := false
	addDevice(context, devices.Devices, baseMessage, isDelta)
	toRemove = getRemoveList(context, devices.Devices)

	if len(toRemove) != 0 {
		removeDevice(context, toRemove, baseMessage, isDelta)
	}
	klog.Info("Deal node detail info successful")
	logrus.WithFields(logrus.Fields{
		"context":    context,
		"devicelist": context.DeviceList,
	}).Infof("dealMembershipDetail finished")
	return nil
}

//func dealTwinMsg(context *dmcontext.DMContext, resource string, msg interface{}) error {
//	message, ok := msg.(*model.Message)
//	if !ok {
//		return errors.New("msg not Message type")
//	}
//
//	contentData, ok := message.Content.([]byte)
//	TwinMsg, err := dmtype.UnmarshalTwinMsg(contentData)
//	if err != nil {
//
//	}
//	logrus.WithFields(logrus.Fields{
//		"module":      "dmservice",
//		"func":        "dealMembershipUpdate()",
//		"message":     message,
//		"contentData": TwinMsg,
//	}).Infof("dealtwin messages")
//	return nil
//}

func dealMembershipGetInner(context *dmcontext.DMContext, payload []byte) error {
	klog.Info("Deal getting membership event")
	result := []byte("")
	edgeGet, err := dmtype.UnmarshalBaseMessage(payload)
	para := dmtype.Parameter{}
	now := time.Now().UnixNano() / 1e6

	if err != nil {
		klog.Errorf("Unmarshal get membership info %s failed , err: %#v", string(payload), err)
		para.Code = dmcommon.BadRequestCode
		para.Reason = fmt.Sprintf("Unmarshal get membership info %s failed , err: %#v", string(payload), err)
		var jsonErr error
		result, jsonErr = dmtype.BuildErrorResult(para)
		if jsonErr != nil {
			klog.Errorf("Unmarshal error result error, err: %v", jsonErr)
		}
	} else {
		para.EventID = edgeGet.EventID
		var devices []*dmtype.Device
		context.DeviceList.Range(func(key interface{}, value interface{}) bool {
			device, ok := value.(*dmtype.Device)
			if !ok {
				klog.Errorf("Content is not Device type.")
			} else {
				devices = append(devices, device)
			}
			return true
		})
		payload, err := dmtype.BuildMembershipGetResult(dmtype.BaseMessage{EventID: edgeGet.EventID, Timestamp: now}, devices)
		if err != nil {
			klog.Errorf("Marshal membership failed while deal get membership ,err: %#v", err)
		} else {
			result = payload
		}
	}
	topic := dmcommon.MemETPrefix + context.NodeName + dmcommon.MemETGetResultSuffix
	klog.Infof("Deal getting membership successful and send the result")

	context.Send("",
		dmcommon.SendToEdge,
		dmcommon.CommModule,
		context.BuildModelMessage(modules.BusGroup, "", topic, messagepkg.OperationPublish, result))
	return nil
}

func addDevice(context *dmcontext.DMContext, toAdd []dmtype.Device, baseMessage dmtype.BaseMessage, delta bool) {
	logrus.WithFields(logrus.Fields{
		"module":      "dmservice",
		"context":     context,
		"toAdd":       toAdd,
		"baseMessage": baseMessage,
		"delta":       delta,
	}).Infof("addDevice")
	if !delta {
		baseMessage.EventID = ""
	}
	for _, device := range toAdd {
		//if device has existed, step out
		deviceInstance, isDeviceExist := context.GetDevice(device.ID)
		if isDeviceExist {
			if delta {
				klog.Errorf("Add device %s failed, has existed", device.ID)
				continue
			}
			UpdateDeviceMeta(context, device.ID, device.Meta)
			//DealDeviceTwin(context, device.ID, baseMessage.EventID, device.Twin, dealType)
			//todo sync twin
			continue
		}
		var deviceMutex sync.Mutex
		context.DeviceMutex.Store(device.ID, &deviceMutex)

		// TODO: what?
		if delta {
			context.Lock(device.ID)
		}

		deviceInstance = &dmtype.Device{ID: device.ID, Name: device.Name, Description: device.Description, State: device.State}
		context.DeviceList.Store(device.ID, deviceInstance)
		logrus.WithFields(logrus.Fields{
			"context": "context",
		}).Infof("Before addDevice db")
		err := WriteDevice2Sql(device)
		//logrus.WithFields(logrus.Fields{
		//	"context": context,
		//	"toAdd":   device.ID,
		//}).Infof("addDevice db Success")
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err":         err,
				"context":     context,
				"toAdd":       toAdd,
				"baseMessage": baseMessage,
				"delta":       delta,
			}).Errorf("addDevice to db failed")
			klog.Errorf("Add device %s failed due to some error ,err: %#v", device.ID, err)
			context.DeviceList.Delete(device.ID)
			context.Unlock(device.ID)
			continue
		}
	}

}

func removeDevice(context *dmcontext.DMContext, toRemove []dmtype.Device, baseMessage dmtype.BaseMessage, delta bool) {
	logrus.WithFields(logrus.Fields{
		"module":      "dmservice",
		"context":     context,
		"toRemove":    toRemove,
		"baseMessage": baseMessage,
		"delta":       delta,
	}).Infof("removeDevice")
	klog.Infof("Begin to remove devices")
	if !delta {
		baseMessage.EventID = ""
	}
	for _, device := range toRemove {
		//update sqlite
		_, deviceExist := context.GetDevice(device.ID)
		if !deviceExist {
			klog.Errorf("Remove device %s failed, not existed", device.ID)
			continue
		}
		if delta {
			context.Lock(device.ID)
		}

		err := DeleteDevice4Sql(device)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err":         err,
				"context":     context,
				"todelete":    device.ID,
				"baseMessage": baseMessage,
				"delta":       delta,
			}).Errorf("delDevice 4 db failed")
		}
		context.DeviceList.Delete(device.ID)
		context.DeviceMutex.Delete(device.ID)
		logrus.WithFields(logrus.Fields{
			"context":  context,
			"todelete": device.ID,
		}).Infof("delDevice Success")
	}
}

func getRemoveList(context *dmcontext.DMContext, devices []dmtype.Device) []dmtype.Device {
	var toRemove []dmtype.Device
	context.DeviceList.Range(func(key interface{}, value interface{}) bool {
		isExist := false
		for _, v := range devices {
			if strings.Compare(v.ID, key.(string)) == 0 {
				isExist = true
				break
			}
		}
		if !isExist {
			toRemove = append(toRemove, dmtype.Device{ID: key.(string)})
		}
		return true
	})
	logrus.WithFields(logrus.Fields{
		"toRemove": toRemove,
	}).Infof("Remove list")
	return toRemove
}

//WriteDevice2Sql impl by membership
func WriteDevice2Sql(device dmtype.Device) error {
	var err error
	adds := make([]dmdatabase.Device, 0)
	adds = append(adds, dmdatabase.Device{
		ID:          device.ID,
		Name:        device.Name,
		Description: device.Description,
		State:       device.State,
		LastOnline:  device.LastOnline,
	})
	logrus.WithFields(logrus.Fields{
		"context": "context",
		"adds":    adds,
	}).Infof("before add transaction")
	for i := 1; i <= dmcommon.RetryTimes; i++ {
		err = dmdatabase.AddDeviceTrans(adds)
		if err == nil {
			logrus.WithFields(logrus.Fields{
				"adds": "adds",
			}).Infof("add transaction no err")
			break
		}
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Errorf("Add transaction err")
		time.Sleep(dmcommon.RetryInterval)
	}
	return err
}

// DeleteDevice4Sql delete device from sql
func DeleteDevice4Sql(device dmtype.Device) error {
	var err error
	deletes := make([]string, 0)
	deletes = append(deletes, device.ID)
	for i := 1; i <= dmcommon.RetryTimes; i++ {
		err = dmdatabase.DeleteDeviceTrans(deletes)
		if err == nil {
			break
		}
		time.Sleep(dmcommon.RetryInterval)
	}
	return err
}
