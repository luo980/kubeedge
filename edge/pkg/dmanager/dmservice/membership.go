package dmservice

import (
	"errors"
	"fmt"
	"github.com/kubeedge/beehive/pkg/core/model"
	messagepkg "github.com/kubeedge/kubeedge/edge/pkg/common/message"
	"github.com/kubeedge/kubeedge/edge/pkg/common/modules"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmcommon"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmcontext"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmtype"
	"k8s.io/klog/v2"
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

	updateEdgeGroups, err := dmtype.UnmarshalMembershipUpdate(contentData)

	if err != nil {
		klog.Errorf("Unmarshal membership info failed , err: %#v", err)
		return err
	}

	baseMessage := dmtype.BaseMessage{EventID: updateEdgeGroups.EventID}
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
	return nil
}

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
}

func removeDevice(context *dmcontext.DMContext, toRemove []dmtype.Device, baseMessage dmtype.BaseMessage, delta bool) {
}

func getRemoveList(context *dmcontext.DMContext, devices []dmtype.Device) []dmtype.Device {
	return nil
}
