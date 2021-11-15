package cdmanager

import (
	beehiveContext "github.com/kubeedge/beehive/pkg/core/context"
	"github.com/kubeedge/beehive/pkg/core/model"
	"github.com/sirupsen/logrus"
	"k8s.io/klog/v2"
)

func (cdm *cDmanager) runCDManager() {
	go func() {
		for {
			select {
			case <-beehiveContext.Done():
				klog.Warning("Stop DeviceManager Receiving.")
				logrus.WithFields(logrus.Fields{
					"module": "dmanager",
					"func":   "runDeviceManager()",
				}).Infof("Stop DeviceManager Receiving.")

				return
			default:
			}
			if msg, ok := beehiveContext.Receive("cdmgr"); ok == nil {
				klog.Info("DeviceManager receive msg")
				logrus.WithFields(logrus.Fields{
					"msg":    msg,
					"module": "cdmanager",
					"func":   "runcCDeviceManager()",
				}).Infof("CDeviceManager receive msg.")

				err := cdm.distributeMsg(msg)
				if err != nil {
					klog.Info("distributeMsg failed.")
					logrus.WithFields(logrus.Fields{
						"module": "cdmanager",
						"func":   "runCDeviceManager()",
					}).Infof("distributeMsg failed.")

				}
			}
		}
	}()
}

func (cdm *cDmanager) distributeMsg(m interface{}) error {
	klog.Info("cDmanager Received message: %v", m)
	msg, ok := m.(model.Message)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"message": msg,
		}).Infof("Cloud DManager Got Message.")
	}
	return nil
}
