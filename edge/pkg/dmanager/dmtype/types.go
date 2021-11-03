package dmtype

import (
	"encoding/json"
	"errors"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmcommon"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmdatabase"
	wlog "github.com/sirupsen/logrus"
	"time"
	//"time"
)

var log = wlog.New()

//BaseMessage the base struct of event message
type BaseMessage struct {
	EventID   string `json:"event_id"`
	Timestamp int64  `json:"timestamp"`
}

//Device the struct of device
type Device struct {
	ID          string              `json:"id,omitempty"`
	Name        string              `json:"name,omitempty"`
	Description string              `json:"description,omitempty"`
	State       string              `json:"state,omitempty"`
	LastOnline  string              `json:"last_online,omitempty"`
	Meta        map[string]*DevMeta `json:"meta,omitempty"`
	Data        map[string]*DevData `json:"data,omitempty"`
}

//DevMeta the struct of device meta
type DevMeta struct {
	Value     string `json:"value,omitempty"`
	Metadata  string `json:"metadata,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

//DevData the struct of device data
type DevData struct {
	Value     string `json:"value,omitempty"`
	Twin      string `json:"twin,omitempty"`
	Metadata  string `json:"metadata,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

//Parameter container para
type Parameter struct {
	EventID string
	Code    int
	Reason  string
}

// Result the struct of Result for sending
type Result struct {
	BaseMessage
	Code   int    `json:"code,omitempty"`
	Reason string `json:"reason,omitempty"`
}

//MembershipUpdate the struct of membership update
type MembershipUpdate struct {
	BaseMessage
	AddDevices    []Device `json:"added_devices"`
	RemoveDevices []Device `json:"removed_devices"`
}

//MembershipDetail the struct of membership detail
type MembershipDetail struct {
	BaseMessage
	Devices []Device `json:"devices"`
}

//DealAttrResult the result of dealing attr
type DealAttrResult struct {
	Add    []dmdatabase.Device
	Delete []dmdatabase.DeviceDelete
	Update []dmdatabase.DeviceAttrUpdate
	Result map[string]*DevMeta
	Err    error
}

// UnmarshalDeviceDataUpdate unmarshal device twin update
func UnmarshalDeviceDataUpdate(payload []byte) (*DeviceDataUpdate, error) {
	var olddeviceDataUpdate OldDeviceDataUpdate
	var deviceDataUpdate DeviceDataUpdate
	err := json.Unmarshal(payload, &olddeviceDataUpdate)
	deviceDataUpdate = fromOld2New(olddeviceDataUpdate)
	if err != nil {
		return &deviceDataUpdate, ErrorUnmarshal
	}
	if deviceDataUpdate.Data == nil {
		return &deviceDataUpdate, ErrorUpdate
	}

	for key, value := range deviceDataUpdate.Data {
		match := dmcommon.ValidateDataKey(key)
		log.WithFields(wlog.Fields{
			"where":   "types.go",
			"key":     key,
			"value":   value,
			"ismatch": match,
		}).Infof("Unmarshal Device Data format")

		if !match {
			return &deviceDataUpdate, ErrorKey
		}
	}
	return &deviceDataUpdate, nil
}

//func fromOld2New(new map[string]*DevData, old map[string]*MsgTwin) {
func fromOld2New(old OldDeviceDataUpdate) DeviceDataUpdate {

	temp := make(map[string]*DevData)
	//temp := make(map[string]*DevData)
	for k, v := range old.Twin {
		log.WithFields(wlog.Fields{
			"key":   k,
			"Value": v.Actual.Value,
			//"Expected": v.Expected.Value,
			"Metadata": v.Metadata.Type,
		}).Infof("test")
		temp[k] = &DevData{
			Value:     *v.Actual.Value,
			Metadata:  v.Metadata.Type,
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		}
	}
	return DeviceDataUpdate{
		BaseMessage: old.BaseMessage,
		Data:        temp,
	}
}

//DeviceDataUpdate the struct of device data update
type DeviceDataUpdate struct {
	BaseMessage
	Data map[string]*DevData `json:"data"`
}
type OldDeviceDataUpdate struct {
	BaseMessage
	Twin map[string]*MsgTwin `json:"twin"`
}

//MsgTwin the struct of device twin
type MsgTwin struct {
	Expected        *TwinValue    `json:"expected,omitempty"`
	Actual          *TwinValue    `json:"actual,omitempty"`
	Optional        *bool         `json:"optional,omitempty"`
	Metadata        *TypeMetadata `json:"metadata,omitempty"`
	ExpectedVersion *TwinVersion  `json:"expected_version,omitempty"`
	ActualVersion   *TwinVersion  `json:"actual_version,omitempty"`
}

//TwinValue the struct of twin value
type TwinValue struct {
	Value    *string        `json:"value,omitempty"`
	Metadata *ValueMetadata `json:"metadata,omitempty"`
}

//ValueMetadata the meta of value
type ValueMetadata struct {
	Timestamp int64 `json:"timestamp,omitempty"`
}

//TypeMetadata the meta of value type
type TypeMetadata struct {
	Type string `json:"type,omitempty"`
}

//TwinVersion twin version
type TwinVersion struct {
	CloudVersion int64 `json:"cloud"`
	EdgeVersion  int64 `json:"edge"`
}

var ErrorUnmarshal = errors.New("Unmarshal update request body failed, please check the request")
var ErrorUpdate = errors.New("Update twin error, key:twin does not exist")
var ErrorKey = errors.New("The key of twin must only include upper or lowercase letters, number, english, and special letter - _ . , : / @ # and the length of key should be less than 128 bytes")
var ErrorValue = errors.New("The value of twin must only include upper or lowercase letters, number, english, and special letter - _ . , : / @ # and the length of value should be less than 512 bytes")
