package dmanager

import (
	"github.com/kubeedge/beehive/pkg/core"
	"github.com/kubeedge/kubeedge/edge/pkg/common/dbm"
	"github.com/kubeedge/kubeedge/edge/pkg/common/modules"
	dmanagerconfig "github.com/kubeedge/kubeedge/edge/pkg/dmanager/config"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmcontext"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmdatabase"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmmodule"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmtype"
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/edgecore/v1alpha1"
	"github.com/sirupsen/logrus"
	"k8s.io/klog/v2"
	"os"
	"sync"
)

var f, _ = os.OpenFile("/home/luo980/logdir/test.log", os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_APPEND, 0755)

type DManager struct {
	HeartBeatToModule map[string]chan interface{}
	DMContexts        *dmcontext.DMContext
	DMModules         map[string]dmmodule.DMModule
	enable            bool
}

func newDeviceManager(enable bool) *DManager {
	return &DManager{
		HeartBeatToModule: make(map[string]chan interface{}),
		DMModules:         make(map[string]dmmodule.DMModule),
		enable:            enable,
	}
}

func Register(dManager *v1alpha1.DManager, nodeName string) {
	dmanagerconfig.InitConfigure(dManager, nodeName)
	dm := newDeviceManager(dManager.Enable)
	dmdatabase.InitDBTable(dm)
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
	logrus.SetOutput(f)
	//if err != nil {
	//	klog.Errorf("Start DManager Failed, Sync Sqlite error:%v", err)
	//	return
	//}
	dmcontexts, _ := dmcontext.InitDMContext()
	dm.DMContexts = dmcontexts
	klog.Infof("DManager Start Here!")
	logrus.WithFields(logrus.Fields{
		"module": "dmanager",
		"func":   "Start()",
	}).Infof("DManager Start Here!")
	devices := new([]dmdatabase.Device)
	_, err := dbm.DBAccess.QueryTable("device").All(devices)

	for _, device := range *devices {
		_, exist := dmcontexts.GetDevice(device.ID)
		if !exist {
			var deviceMutex sync.Mutex
			dmcontexts.DeviceMutex.Store(device.ID, &deviceMutex)
		}
		dm.DMContexts.DeviceList.Store(device.ID, dmtype.Device{
			ID:          device.ID,
			Name:        device.Name,
			Description: device.Description,
			State:       device.State,
			LastOnline:  device.LastOnline,
			Meta:        nil,
			Data:        nil,
		})
		logrus.Error("DeviceList now: ", dm.DMContexts.DeviceList)
	}

	if err != nil {
		print(err)
	}

	klog.Infof("DManager Sync!")
	logrus.WithFields(logrus.Fields{
		"module":     "sync from sqlite",
		"DeviceList": dm.DMContexts.DeviceList,
	}).Infof("DManager DeviceList Sync!")

	dm.runDeviceManager()

}
