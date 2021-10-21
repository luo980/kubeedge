package dmtype

import (
	"encoding/json"
	"k8s.io/klog/v2"
)

//DeviceUpdate device update
type DeviceUpdate struct {
	BaseMessage
	State      string              `json:"state,omitempty"`
	Attributes map[string]*MsgAttr `json:"attributes"`
}

//UnmarshalDeviceUpdate unmarshal device update
func UnmarshalDeviceUpdate(payload []byte) (*DeviceUpdate, error) {
	var get DeviceUpdate
	err := json.Unmarshal(payload, &get)
	klog.Infof("Unmarshal json outcome is %s", get)
	if err != nil {
		return nil, err
	}
	return &get, nil
}
