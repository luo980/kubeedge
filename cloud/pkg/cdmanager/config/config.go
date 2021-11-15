package config

import (
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/cloudcore/v1alpha1"
	"sync"
)

var Config Configure
var once sync.Once

type Configure struct {
	v1alpha1.CDmanager
}

func InitConfigure(cd *v1alpha1.CDmanager) {
	once.Do(func() {
		Config = Configure{
			CDmanager: *cd,
		}
	})
}
