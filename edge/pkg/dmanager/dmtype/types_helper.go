package dmtype

import (
	"encoding/json"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmdatabase"
	"k8s.io/klog/v2"
	"time"
)

//DeviceUpdate device update
type DeviceUpdate struct {
	BaseMessage
	State      string              `json:"state,omitempty"`
	Attributes map[string]*MsgAttr `json:"attributes"`
}

//MembershipGetResult membership get result
type MembershipGetResult struct {
	BaseMessage
	Devices []Device `json:"devices"`
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

// BuildErrorResult build error result
func BuildErrorResult(para Parameter) ([]byte, error) {
	result := Result{BaseMessage: BaseMessage{Timestamp: time.Now().UnixNano() / 1e6,
		EventID: para.EventID},
		Code:   para.Code,
		Reason: para.Reason}
	errorResult, err := json.Marshal(result)
	if err != nil {
		return []byte(""), err
	}
	return errorResult, nil
}

//UnmarshalBaseMessage Unmarshal get
func UnmarshalBaseMessage(payload []byte) (*BaseMessage, error) {
	var get BaseMessage
	err := json.Unmarshal(payload, &get)
	if err != nil {
		return nil, err
	}
	return &get, nil
}

//BuildMembershipGetResult build memebership
func BuildMembershipGetResult(baseMessage BaseMessage, devices []*Device) ([]byte, error) {
	result := make([]Device, 0, len(devices))
	for _, v := range devices {
		result = append(result, Device{
			ID:          v.ID,
			Name:        v.Name,
			Description: v.Description,
			State:       v.State,
			LastOnline:  v.LastOnline,
			Attributes:  v.Attributes})
	}
	payload, err := json.Marshal(MembershipGetResult{BaseMessage: baseMessage, Devices: result})
	if err != nil {
		return []byte(""), err
	}
	return payload, nil
}

//UnmarshalMembershipUpdate Unmarshal membershipupdate
func UnmarshalMembershipUpdate(payload []byte) (*MembershipUpdate, error) {
	var membershipUpdate MembershipUpdate
	err := json.Unmarshal(payload, &membershipUpdate)
	if err != nil {
		return nil, err
	}
	return &membershipUpdate, nil
}

//UnmarshalMembershipDetail Unmarshal membershipdetail
func UnmarshalMembershipDetail(payload []byte) (*MembershipDetail, error) {
	var membershipDetail MembershipDetail
	err := json.Unmarshal(payload, &membershipDetail)
	if err != nil {
		return nil, err
	}
	return &membershipDetail, nil
}

//UnmarshalTwinMsg Unmarshal TwinMsg
func UnmarshalTwinMsg(payload []byte) (*TwinMsg, error) {
	var twinmsg TwinMsg
	err := json.Unmarshal(payload, &twinmsg)
	if err != nil {
		return nil, err
	}
	return &twinmsg, nil
}

//CopyMsgAttr copy msg attr
func CopyMsgAttr(msgAttr *MsgAttr) MsgAttr {
	var result MsgAttr
	payload, _ := json.Marshal(msgAttr)
	json.Unmarshal(payload, &result)
	return result
}

////MsgAttrToDeviceAttr msgattr to deviceattr
func MsgAttrToDeviceAttr(name string, msgAttr *MsgAttr) dmdatabase.Device {
	//	attrType := "string"
	//	if msgAttr.Metadata != nil {
	//		attrType = msgAttr.Metadata.Type
	//	}
	//	optional := true
	//	if msgAttr.Optional != nil {
	//		optional = *msgAttr.Optional
	//	}
	//	return dtclient.DeviceAttr{
	//		Name:     name,
	//		AttrType: attrType,
	//		Optional: optional}
	return dmdatabase.Device{}
}
