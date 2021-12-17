package cdmanager

import (
	"github.com/kubeedge/beehive/pkg/core"
	beehiveContext "github.com/kubeedge/beehive/pkg/core/context"
	"github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/cdmcontext"
	"github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/cdmmodule"
	"github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/common"
	cdmconfig "github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/config"
	"github.com/kubeedge/kubeedge/cloud/pkg/common/modules"
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/cloudcore/v1alpha1"
	"github.com/sirupsen/logrus"
	"os"
)

var f, _ = os.OpenFile("/home/luo980/logdir/ctest.log", os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_APPEND, 0755)

type CDManager struct {
	CDMContexts		*cdmcontext.CDMContext
	CDModules		map[string]cdmmodule.CDMModule
	enable      	bool
}

func Register(cdm *v1alpha1.CDmanager) {
	cdmconfig.InitConfigure(cdm)
	core.Register(newCDManager(cdm.Enable))
}

func (cdm *CDManager) Name() string {
	return modules.CDmamagerModuleName
}

func (cdm *CDManager) Group() string {
	return modules.CDmanagerGroupName
}

// Enable indicates whether enable this module
func (cdm *CDManager) Enable() bool {
	return cdm.enable
}

func newCDManager(enable bool) *CDManager {
	return &CDManager{
		CDModules:	make(map[string]cdmmodule.CDMModule),
		enable: 	enable,
	}
}

func (cdm *CDManager) Start() {
	logrus.SetOutput(f)

	cdmContexts, _ := cdmcontext.InitCDMContext()
	cdm.CDMContexts = cdmContexts

	logrus.WithFields(logrus.Fields{
		"Hello": "Enter cdmanager module.",
	}).Infof("Start CDManager")

	cdm.runCDManager()
}

func (cdm *CDManager) runCDManager(){
	moduleNames := []string{common.CommModule, common.DeviceModule, common.ServerModule, common.EdgeModule}
	for _, v := range moduleNames{
		cdm.RegisterDMModule(v)
		go cdm.CDModules[v].Start()
	}
	go func(){
		for {
			select {
			case <- beehiveContext.Done():
				logrus.Warning("Cloud DeviceManager stop receiving.")
				return
			default:
			}
			if msg, ok := beehiveContext.Receive("cdmgr"); ok == nil{
				logrus.Info("Cloud DeviceManager receive msg: ", msg)
			}
		}
	}()
}

//RegisterCDMModule register cdmmodule
func (cdm CDManager) RegisterDMModule(name string) {
	module := cdmmodule.CDMModule{
		Name: name,
	}

	cdm.CDMContexts.CommChan[name] = make(chan interface{}, 128)
	module.InitWorker(cdm.CDMContexts.CommChan[name], cdm.CDMContexts.ConfirmChan,  cdm.CDMContexts)
	cdm.CDModules[name] = module
}