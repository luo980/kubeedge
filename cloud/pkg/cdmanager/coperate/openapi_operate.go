package coperate

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1alpha2device "github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/cdeviceapi/v1alpha2"
)

func CreateNewDevice(recv_msg RegDevice) v1alpha2device.Device{
	DLabels := make(map[string]string)
	DLabels["DeviceID"] = recv_msg.DeviceID
	DLabels["MAC"] = recv_msg.MAC
	DLabels["Location"] = recv_msg.Location
	DLabels["JoinTime"] = recv_msg.JoinTime
	DLabels["Manufacturer"] = recv_msg.Manufacturer

	DModel :=  v1alpha2device.DeviceSpec{}


	MatchE := v1.LocalObjectReference{
		Name: recv_msg.DeviceModel,
	}

	selectorValue := make([]string, 1)
	selectorValue[0] = recv_msg.EdgeName

	NodeSTR := make([]v1.NodeSelectorRequirement, 1)
	NodeSTR[0] = v1.NodeSelectorRequirement{
		Key: 		"",
		Operator: 	"In",
		Values: 	selectorValue,
	}

	NodeST := make([]v1.NodeSelectorTerm, 1)
	NodeST[0] = v1.NodeSelectorTerm{
		MatchExpressions:	NodeSTR,
	}

	NodeS := v1.NodeSelector{
		NodeSelectorTerms:	NodeST,
	}

	DModel.DeviceModelRef = &MatchE
	DModel.NodeSelector = &NodeS

	typestring := make(map[string]string)
	typestring["type"] = "string"

	DTwin := make([]v1alpha2device.Twin, 1)
	//TP_temperature := v1alpha2device.TwinProperty{}
	DTwin[0].PropertyName = "temperature"
	DTwin[0].Desired.Metadata = typestring
	DTwin[0].Desired.Value = ""

	var DStatus v1alpha2device.DeviceStatus
	DStatus.Twins = DTwin

	newDevice := v1alpha2device.Device{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Device",
			APIVersion: "devices.kubeedge.io/v1alpha2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      recv_msg.DeviceName,
			Namespace: "default",
			Labels:    DLabels,
		},
		Spec:   DModel,
		Status: DStatus,
	}
	return newDevice
}