package config

import (
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/edgecore/v1alpha1"
	"sync"
)

var config Configure
var once sync.Once

type Configure struct {
	v1alpha1.DManager
	NodeName string
}

func InitConfigure(deviceManager *v1alpha1.DManager, nodeName string) {
	once.Do(func() {
		config = Configure{
			DManager: *deviceManager,
			NodeName: nodeName,
		}
	})
}

func Get() *Configure {
	return &config
}
