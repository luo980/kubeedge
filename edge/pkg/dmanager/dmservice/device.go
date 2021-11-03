package dmservice

import (
	"errors"
	"github.com/kubeedge/beehive/pkg/core/model"
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
					log.WithFields(wlog.Fields{
						"dmMsg.Action": dmMsg.Action,
						"fn":           fn,
						"func":         "Start()",
					}).Infof("Device Module Received Msg")
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
	msg, err := dmtype.UnmarshalDeviceDataUpdate(payload)
	log.WithFields(wlog.Fields{
		"where":  "Updated",
		"result": result,
		"msg":    msg,
		"err":    err,
	}).Infof("unmarshal finished")
	eventID := msg.EventID
	dealDeviceDiffUpdate(context, deviceID, eventID, msg.Data)

	return nil
}

func dealDeviceDiffUpdate(context *dmcontext.DMContext, deviceID string, eventID string, data map[string]*dmtype.DevData) error {
	//now := time.Now().UnixNano() / 1e6
	device, isExist := context.GetDevice(deviceID)
	log.WithFields(wlog.Fields{
		"deviceID": deviceID,
		"device":   device,
		"where":    "dealDeviceDiffUpdate",
	}).Infof("GetDevice")
	if !isExist {
		return errors.New("no device")
	}
	if data == nil {
		return nil
	}
	oDeviceData := device.Data
	for key, value := range data {
		if innerData, exist := oDeviceData[key]; exist {
			if value.Metadata == "" {
				log.WithFields(wlog.Fields{
					"where": "dealDeviceDiffUpdate",
				}).Errorf("datas' Meta lost")
			}
			if data == nil || strings.Compare(value.Metadata, "deleted") == 0 {
				delete(oDeviceData, key)
				continue
			}
			oDeviceData[key] = innerData
		} else {
			//append(oDeviceData[key], value)
		}
	}
	log.WithFields(wlog.Fields{
		"where": "dealDeviceDiffUpdate",
		"what":  oDeviceData,
	}).Errorf("device data")
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
	}).Info("here to handle device state")

	// state refers to definition in mappers-go/pkg/common/const.go
	state := strings.ToLower(updatedDevice.State)
	switch state {
	case "online", "offline", "ok", "unknown", "disconnected":
	default:
		return nil
	}
	device.State = state
	device.LastOnline = time.Now().Format("2006-01-02 15:04:05")

	return nil
}

func UpdateDeviceMeta(context *dmcontext.DMContext, deviceID string, devMeta map[string]*dmtype.DevMeta) (interface{}, error) {
	log.WithFields(wlog.Fields{
		"where": "device.go",
		"what":  devMeta,
	}).Info("here to update device info")
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
