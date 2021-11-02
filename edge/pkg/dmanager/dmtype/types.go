package dmtype

import (
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmdatabase"
)

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
