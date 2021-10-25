package dmcontext

import (
	"context"
	"errors"
	"github.com/kubeedge/beehive/pkg/core/model"
	"github.com/kubeedge/kubeedge/edge/pkg/common/modules"
	dmanagerconfig "github.com/kubeedge/kubeedge/edge/pkg/dmanager/config"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmcommon"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmtype"
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

		ModelList:  &sync.Map{},
		ModelMutex: &sync.Map{},

		AbilityList:  &sync.Map{},
		AbilityMutex: &sync.Map{},

		Mutex: &sync.RWMutex{},
		State: dmcommon.Disconnected,
	}, nil
}

//GetDevice get device
func (dmc *DMContext) GetDevice(deviceID string) (*dmtype.Device, bool) {
	d, ok := dmc.DeviceList.Load(deviceID)
	if ok {
		if device, isDevice := d.(*dmtype.Device); isDevice {
			return device, true
		}
		return nil, false
	}
	return nil, false
}

//CommTo communicate
func (dmc *DMContext) CommTo(dmcName string, content interface{}) error {
	if v, exist := dmc.CommChan[dmcName]; exist {
		v <- content
		return nil
	}
	return errors.New("Not found chan to communicate")
}

//Lock get the lock of the device
func (dmc *DMContext) Lock(deviceID string) bool {
	deviceMutex, ok := dmc.GetMutex(deviceID)
	if ok {
		dmc.Mutex.RLock()
		deviceMutex.Lock()
		return true
	}
	return false
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

func (dmc *DMContext) HeartBeat(dmmName string, content interface{}) error {
	return nil
}

//Send send result
func (dmc *DMContext) Send(identity string, action string, module string, msg *model.Message) error {
	dmMsg := &dmtype.DMMessage{
		Action:   action,
		Identity: identity,
		Type:     module,
		Msg:      msg}
	return dmc.CommTo(module, dmMsg)
}

//BuildModelMessage build mode messages
func (dmc *DMContext) BuildModelMessage(group string, parentID string, resource string, operation string, content interface{}) *model.Message {
	msg := model.NewMessage(parentID)
	msg.BuildRouter(modules.TwinGroup, group, resource, operation)
	msg.Content = content
	return msg
}
