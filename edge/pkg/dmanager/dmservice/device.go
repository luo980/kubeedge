package dmservice

import (
	"errors"
	"github.com/kubeedge/beehive/pkg/core/model"
	"github.com/kubeedge/kubeedge/edge/pkg/devicetwin/dtcontext"
	"github.com/kubeedge/kubeedge/edge/pkg/devicetwin/dttype"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmcommon"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmcontext"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmtype"
	"k8s.io/klog/v2"
	"strings"
	"time"

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
	deviceActionCallBack[dmcommon.TwinUpdate] = dealDeviceDataUpdate

}

func dealDeviceDataUpdate(context *dmcontext.DMContext, resource string, msg interface{}) error {
	log.WithFields(wlog.Fields{
		"where": "device.go",
	}).Info("here to update device data")
	message, ok := msg.(*model.Message)
	if !ok {
		log.WithFields(wlog.Fields{
			"where": "dealDeviceDataUpdate",
		}).Errorf("Not message")
	}
	content, ok := message.Content.([]byte)
	if !ok {
		log.WithFields(wlog.Fields{
			"where": "dealDeviceDataUpdate",
		}).Errorf("Not Device data message")
	}
	context.Lock(resource)
	Updated(context, resource, content)
	context.Unlock(resource)
	return nil
}

func Updated(context *dmcontext.DMContext, deviceID string, payload []byte) error {
	result := []byte("")
	msg, err := dmtype.UnmarshalDeviceTwinUpdate(payload)

	return nil
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

func UpdateDeviceMeta(context *dmcontext.DMContext, deviceID string, devMeta map[string]*dmtype.DevMeta, baseMessage dmtype.BaseMessage, dealType bool) (interface{}, error) {
	log.WithFields(wlog.Fields{
		"where": "device.go",
		"what":  devMeta,
	}).Info("here to update device info")
	doc, docExist := context.DeviceList.Load(deviceID)
	if !docExist {
		log.WithFields(wlog.Fields{
			"where": "device.go",
		}).Info("Device not exist in DeviceList")
		return nil, nil
	}
	var err error
	DealMsgMeta(context, deviceID, devMeta)

	return nil, nil
}

func DealMsgMeta(context *dmcontext.DMContext, deviceID string, meta map[string]*dmtype.DevMeta) {
	device, _ := context.GetDevice(deviceID)
	result := make(map[string]*dmtype.DevMeta)
	if device.Meta == nil {
		device.Meta = make(map[string]*dmtype.DevMeta)
	}
	devmetas := device.Meta
	for key, metas := range meta {
		// if key exists
		if devmeta, exist := devmetas[key]; exist {
			// nil meta
			if metas == nil {
				result[key] = nil
				delete(devmetas, key)
				continue
			}
			// not nil meta & update
			result[key] = &dmtype.DevMeta{}
			if strings.Compare(devmeta.Value, metas.Value) != 0 {
				devmeta.Value = metas.Value
			}
			// key not exists
		} else {
			if len(metas.Metadata) != 0 {
				devmeta.Metadata = metas.Metadata
			}
			devmeta.Value = metas.Value
		}
	}
}
