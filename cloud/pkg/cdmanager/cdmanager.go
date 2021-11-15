package cdmanager

import (
	"github.com/kubeedge/beehive/pkg/core"
	"github.com/kubeedge/beehive/pkg/core/model"
	cdmconfig "github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/config"
	"github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/server"
	"github.com/kubeedge/kubeedge/cloud/pkg/common/modules"
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/cloudcore/v1alpha1"
	"github.com/sirupsen/logrus"
	"os"
)

var f, _ = os.OpenFile("/home/luo980/logdir/ctest.log", os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_APPEND, 0755)

type cDmanager struct {
	//downstream *controller.DownstreamController
	//upstream   *controller.UpstreamController
	queryserver *server.InteractServer
	messageq    chan model.Message
	enable      bool
}

func Register(cdm *v1alpha1.CDmanager) {
	cdmconfig.InitConfigure(cdm)
	core.Register(newCDmanager(cdm.Enable))
}

func (cdm *cDmanager) Name() string {
	return modules.CDmamagerModuleName
}

func (cdm *cDmanager) Group() string {
	return modules.CDmanagerGroupName
}

// Enable indicates whether enable this module
func (cdm *cDmanager) Enable() bool {
	return cdm.enable
}

func newCDmanager(enable bool) *cDmanager {

	return &cDmanager{
		//queryserver: queryserver,
		enable: enable,
	}
}

func (cdm *cDmanager) Start() {
	logrus.SetOutput(f)
	logrus.WithFields(logrus.Fields{
		"Hello": "Enter cdmanager module.",
	}).Infof("Start CDManager")
	cdm.runCDManager()
}
