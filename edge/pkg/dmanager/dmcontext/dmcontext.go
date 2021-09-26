package dmcontext

import (
	"context"
	"errors"
	"github.com/kubeedge/kubeedge/edge/pkg/devicetwin/dtcommon"
	dmanagerconfig "github.com/kubeedge/kubeedge/edge/pkg/dmanager/config"
	"k8s.io/klog/v2"
	"sync"
)

type DMContext struct {
	GroupID       string
	NodeName      string
	CommChan      map[string]chan interface{}
	ConfirmChan   chan interface{}
	ModulesHealth *sync.Map

	//returns deadline
	ModulesContext *context.Context

	ConfirmMap  *sync.Map
	DeviceList  *sync.Map
	DeviceMutex *sync.Map

	ModelList    *sync.Map
	ModelMutex   *sync.Map
	AbilityList  *sync.Map
	AbilityMutex *sync.Map
	Mutex        *sync.RWMutex
	State        string
}

//InitDMContext init dmcontext
func InitDMContext() (*DMContext, error) {
	return &DMContext{
		GroupID:       "",
		NodeName:      dmanagerconfig.Get().NodeName,
		CommChan:      make(map[string]chan interface{}),
		ConfirmChan:   make(chan interface{}, 1000),
		ModulesHealth: &sync.Map{},

		ConfirmMap:  &sync.Map{},
		DeviceList:  &sync.Map{},
		DeviceMutex: &sync.Map{},

		ModelList:    &sync.Map{},
		ModelMutex:   &sync.Map{},
		AbilityList:  &sync.Map{},
		AbilityMutex: &sync.Map{},

		Mutex: &sync.RWMutex{},
		State: dtcommon.Disconnected,
	}, nil
}

//CommTo communicate
func (dmc *DMContext) CommTo(dmcName string, content interface{}) error {
	if v, exist := dmc.CommChan[dmcName]; exist {
		v <- content
		return nil
	}
	return errors.New("Not found chan to communicate")
}

//Unlock remove the lock of the device
func (dmc *DMContext) Unlock(deviceID string) bool {
	deviceMutex, ok := dmc.GetMutex(deviceID)
	if ok {
		deviceMutex.Unlock()
		dmc.Mutex.RUnlock()
		return true
	}
	return false
}

//GetMutex get mutex
func (dmc *DMContext) GetMutex(deviceID string) (*sync.Mutex, bool) {
	v, mutexExist := dmc.DeviceMutex.Load(deviceID)
	if !mutexExist {
		klog.Errorf("GetMutex device %s not exist", deviceID)
		return nil, false
	}
	mutex, isMutex := v.(*sync.Mutex)
	if isMutex {
		return mutex, true
	}
	return nil, false
}

// LockAll get all lock
func (dmc *DMContext) LockAll() {
	dmc.Mutex.Lock()
}

// UnlockAll get all lock
func (dmc *DMContext) UnlockAll() {
	dmc.Mutex.Unlock()
}

//IsDeviceExist judge device is exist
func (dmc *DMContext) IsDeviceExist(deviceID string) bool {
	_, ok := dmc.DeviceList.Load(deviceID)
	return ok
}

//IsModelExist judge model is exist
func (dmc *DMContext) IsModelExist(modelID string) bool {
	_, ok := dmc.ModelList.Load(modelID)
	return ok
}

//IsAbilityExist judge ability is exist
func (dmc *DMContext) IsAbilityExist(abilityID string) bool {
	_, ok := dmc.AbilityList.Load(abilityID)
	return ok
}
