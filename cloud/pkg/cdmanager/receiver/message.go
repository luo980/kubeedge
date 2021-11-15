package receiver

import (
	beehiveContext "github.com/kubeedge/beehive/pkg/core/context"
	"github.com/kubeedge/beehive/pkg/core/model"
	"github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/constants"
	"k8s.io/klog/v2"
)

type DeviceReceiver struct {
	messageLayer MessageLayer
	statusChan   chan model.Message
}

func (dr *DeviceReceiver) Start() error {
	dr.statusChan = make(chan model.Message)
	return nil
}

func (dr *DeviceReceiver) dispatchMessage() {
	for {
		select {
		case <-beehiveContext.Done():
			klog.Info("Stop dispatchMessage")
			return
		default:
		}
		msg, err := dr.messageLayer.Receive()
		if err != nil {
			continue
		}
		resourceType, err := GetResourceType(msg.GetResource())
		switch resourceType {
		case constants.ResourceTypeDataUpdate:
			dr.statusChan <- msg
		case constants.ResourceTypeMembershipDetail:
		default:
			klog.Warningf("Message: %s, with resource type: %s not intended for device controller", msg.GetID(), resourceType)
		}

	}
}
