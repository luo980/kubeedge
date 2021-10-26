package dmservice

import (
	"errors"
	"github.com/kubeedge/beehive/pkg/core/model"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmcommon"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmcontext"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmtype"
	"k8s.io/klog/v2"
	"strings"

	wlog "github.com/sirupsen/logrus"
)

var log = wlog.New()

type DeviceWorker struct {
	Worker
	Group string
}

var (
	//deviceActionCallBack map for action to callback
	deviceActionCallBack map[string]CallBack
)

func (dw DeviceWorker) Start() {
	initDeviceActionCallBack()
	for {
		select {
		case msg, ok := <-dw.ReceiverChan:
			if !ok {
				return
			}
			if dmMsg, isDMMessage := msg.(*dmtype.DMMessage); isDMMessage {
				if fn, exist := deviceActionCallBack[dmMsg.Action]; exist {
					err := fn(dw.DMContexts, dmMsg.Identity, dmMsg.Msg)
					if err != nil {
						klog.Errorf("DeviceModule deal %s event failed: %v", dmMsg.Action, err)
					}
				}
			}
		case v, ok := <-dw.HeartBeatChan:
			if !ok {
				return
			}
			if err := dw.DMContexts.HeartBeat(dw.Group, v); err != nil {
				klog.Infof("Heartbeat err is : %s", err)
				return
			}
		}
	}
}

func initDeviceActionCallBack() {
	deviceActionCallBack = make(map[string]CallBack)
	deviceActionCallBack[dmcommon.DeviceUpdated] = dealDeviceAttrUpdate
	deviceActionCallBack[dmcommon.DeviceStateUpdate] = dealDeviceStateUpdate
}

func dealDeviceAttrUpdate(context *dmcontext.DMContext, resource string, msg interface{}) error {
	return nil
}

func dealDeviceStateUpdate(context *dmcontext.DMContext, resource string, msg interface{}) error {
	message, ok := msg.(*model.Message)
	if !ok {
		return errors.New("DeviceState Update Msg is not DMMessage type")
	}

	updatedDevice, err := dmtype.UnmarshalDeviceUpdate(message.Content.([]byte))
	if err != nil {
		klog.Errorf("Unmarshal device info failed, err: %#v", err)
		return err
	}
	deviceID := resource
	defer context.Unlock(deviceID)
	context.Lock(deviceID)
	doc, docExist := context.DeviceList.Load(deviceID)
	if !docExist {
		return nil
	}
	device, ok := doc.(*dmtype.Device)
	log.WithFields(wlog.Fields{
		"where": "device.go",
		"what":  device,
	}).Info("here to debug device")

	// state refers to definition in mappers-go/pkg/common/const.go
	state := strings.ToLower(updatedDevice.State)
	switch state {
	case "online", "offline", "ok", "unknown", "disconnected":
	default:
		return nil
	}

	return nil
}

func UpdateDeviceAttr(context *dmcontext.DMContext, deviceID string, attributes map[string]*dmtype.MsgAttr, baseMessage dmtype.BaseMessage, dealType bool) (interface{}, error) {
	return nil, nil
}

//DealMsgAttr get diff,0:update, 1:detail
//func DealMsgAttr(context *dmcontext.DMContext, deviceID string, msgAttrs map[string]*dmtype.MsgAttr, dealType int) dmtype.DealAttrResult {
//	device, ok := context.GetDevice(deviceID)
//	if !ok {
//		klog.Errorf("Can't Get device %v", deviceID)
//	}
//	attrs := device.Attributes
//	if attrs == nil {
//		device.Attributes = make(map[string]*dmtype.MsgAttr)
//		attrs = device.Attributes
//	}
//	add := make([]dmdatabase.Device, 0)
//	deletes := make([]dmdatabase.DeviceDelete, 0)
//	update := make([]dmdatabase.DeviceAttrUpdate, 0)
//	result := make(map[string]*dmtype.MsgAttr)
//
//	for key, msgAttr := range msgAttrs {
//		if attr, exist := attrs[key]; exist {
//			if msgAttr == nil && dealType == 0 {
//				if *attr.Optional {
//					deletes = append(deletes, dmdatabase.DeviceDelete{DeviceID: deviceID, Name: key})
//					result[key] = nil
//					delete(attrs, key)
//				}
//				continue
//			}
//			isChange := false
//			cols := make(map[string]interface{})
//			result[key] = &dmtype.MsgAttr{}
//			if strings.Compare(attr.Value, msgAttr.Value) != 0 {
//				attr.Value = msgAttr.Value
//
//				cols["value"] = msgAttr.Value
//				result[key].Value = msgAttr.Value
//
//				isChange = true
//			}
//			if msgAttr.Metadata != nil {
//				msgMetaJSON, _ := json.Marshal(msgAttr.Metadata)
//				attrMetaJSON, _ := json.Marshal(attr.Metadata)
//				if strings.Compare(string(msgMetaJSON), string(attrMetaJSON)) != 0 {
//					cols["attr_type"] = msgAttr.Metadata.Type
//					meta := dmtype.CopyMsgAttr(msgAttr)
//					attr.Metadata = meta.Metadata
//					msgAttr.Metadata.Type = ""
//					metaJSON, _ := json.Marshal(msgAttr.Metadata)
//					cols["metadata"] = string(metaJSON)
//					msgAttr.Metadata.Type = cols["attr_type"].(string)
//					result[key].Metadata = meta.Metadata
//					isChange = true
//				}
//			}
//			if msgAttr.Optional != nil {
//				if *msgAttr.Optional != *attr.Optional && *attr.Optional {
//					optional := *msgAttr.Optional
//					cols["optional"] = optional
//					attr.Optional = &optional
//					result[key].Optional = &optional
//					isChange = true
//				}
//			}
//			if isChange {
//				update = append(update, dmdatabase.DeviceAttrUpdate{DeviceID: deviceID, Name: key, Cols: cols})
//			} else {
//				delete(result, key)
//			}
//		} else {
//			deviceAttr := dmtype.MsgAttrToDeviceAttr(key, msgAttr)
//			deviceAttr.DeviceID = deviceID
//			deviceAttr.Value = msgAttr.Value
//			if msgAttr.Optional != nil {
//				optional := *msgAttr.Optional
//				deviceAttr.Optional = optional
//			}
//			if msgAttr.Metadata != nil {
//				//todo
//				deviceAttr.AttrType = msgAttr.Metadata.Type
//				msgAttr.Metadata.Type = ""
//				metaJSON, _ := json.Marshal(msgAttr.Metadata)
//				msgAttr.Metadata.Type = deviceAttr.AttrType
//				deviceAttr.Metadata = string(metaJSON)
//			}
//			add = append(add, deviceAttr)
//			attrs[key] = msgAttr
//			result[key] = msgAttr
//		}
//	}
//	if dealType > 0 {
//		for key := range attrs {
//			if _, exist := msgAttrs[key]; !exist {
//				deletes = append(deletes, dmdatabase.DeviceDelete{DeviceID: deviceID, Name: key})
//				result[key] = nil
//			}
//		}
//		for _, v := range deletes {
//			delete(attrs, v.Name)
//		}
//	}
//	//return dmtype.DealAttrResult{Add: add, Delete: deletes, Update: update, Result: result, Err: nil}
//	return dmtype.DealAttrResult{}
//}
