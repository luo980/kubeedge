package dmanager

import (
	beehiveContext "github.com/kubeedge/beehive/pkg/core/context"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmcommon"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmmodule"
	"k8s.io/klog/v2"
)

func (dm *DManager) runDeviceManager() {
	moduleNames := []string{dmcommon.DeviceModule, dmcommon.MemModule, dmcommon.AbilityModule, dmcommon.CommModule}
	for _, v := range moduleNames {
		dm.RegisterDMModule(v)
		go dm.DMModules[v].Start()
	}
	go func() {
		for {
			select {
			case <-beehiveContext.Done():
				klog.Warning("Stop DeviceManager Receiving.")
				return
			default:
			}
			if msg, ok := beehiveContext.Receive("dmgr"); ok == nil {
				klog.Info("DeviceManager receive msg")
				err := dm.distributeMsg(msg)
				if err != nil {
					klog.Info("distributeMsg failed.")
				}
			}
		}
	}()
}

//RegisterDMModule register dmmodule
func (dm *DManager) RegisterDMModule(name string) {
	module := dmmodule.DMModule{
		Name: name,
	}

	dm.DMContexts.CommChan[name] = make(chan interface{}, 128)
	dm.HeartBeatToModule[name] = make(chan interface{}, 128)
	module.InitWorker(dm.DMContexts.CommChan[name], dm.DMContexts.ConfirmChan, dm.HeartBeatToModule[name], dm.DMContexts)
	dm.DMModules[name] = module
}

// distributeMsg distribute message to different modules
func (dm *DManager) distributeMsg(m interface{}) error {
	klog.Info("Received message: %v", m)
	return nil
}
