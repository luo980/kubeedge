package dmdatabase

import (
	"github.com/astaxie/beego/orm"
	"github.com/kubeedge/beehive/pkg/core"
	"k8s.io/klog/v2"
)

const (
	//DeviceTableName device table
	DeviceTableName = "device"
	//DeviceAttrTableName device table
	DeviceAttrTableName = "device_attr"
)

// InitDBTable create table
func InitDBTable(module core.Module) {
	klog.Infof("Begin to register %v db model", module.Name())

	if !module.Enable() {
		klog.Infof("Module %s is disabled, DB meta for it will not be registered", module.Name())
		return
	}
	orm.RegisterModel(new(Device))
}
