package dmanager

import (
	"github.com/kubeedge/beehive/pkg/core"
	"github.com/kubeedge/kubeedge/edge/pkg/common/modules"
	dmanagerconfig "github.com/kubeedge/kubeedge/edge/pkg/dmanager/config"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmcontext"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmmodule"
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/edgecore/v1alpha1"
	"k8s.io/klog/v2"
)

type DManager struct {
	DMContexts *dmcontext.DMContext
	DMModules  map[string]dmmodule.DMModule
	enable     bool
}

func newDeviceManager(enable bool) *DManager {
	return &DManager{

		enable: enable,
	}
}

func Register(dManager *v1alpha1.DManager, nodeName string) {
	dmanagerconfig.InitConfigure(dManager, nodeName)
	dm := newDeviceManager(dManager.Enable)
	// Register DManager to Core
	core.Register(dm)
}

func (dm *DManager) Name() string {
	return modules.DManagerModuleName
}

func (dm *DManager) Group() string {
	return modules.DMgrGroup
}

func (dm *DManager) Enable() bool {
	return dm.enable
}

func (dm *DManager) Start() {
	//if err != nil {
	//	klog.Errorf("Start DManager Failed, Sync Sqlite error:%v", err)
	//	return
	//}
	klog.Infof("DManager Start Here!")
}
