package dmanager

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	beehiveContext "github.com/kubeedge/beehive/pkg/core/context"
	"github.com/kubeedge/beehive/pkg/core/model"
	"github.com/kubeedge/kubeedge/edge/pkg/common/modules"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmcommon"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmmodule"
	"github.com/kubeedge/kubeedge/edge/pkg/dmanager/dmtype"
	"github.com/sirupsen/logrus"
	"k8s.io/klog/v2"
	"strings"
)

var (
	//EventActionMap map for event to action
	EventActionMap map[string]map[string]string
	//ActionModuleMap map for action to module
	ActionModuleMap map[string]string
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
				logrus.WithFields(logrus.Fields{
					"module": "dmanager",
					"func":   "runDeviceManager()",
				}).Infof("Stop DeviceManager Receiving.")

				return
			default:
			}
			if msg, ok := beehiveContext.Receive("dmgr"); ok == nil {
				klog.Info("DeviceManager receive msg")
				logrus.WithFields(logrus.Fields{
					"msg":    msg,
					"module": "dmanager",
					"func":   "runDeviceManager()",
				}).Infof("DeviceManager receive msg.")

				err := dm.distributeMsg(msg)
				if err != nil {
					klog.Info("distributeMsg failed.")
					logrus.WithFields(logrus.Fields{
						"module": "dmanager",
						"func":   "runDeviceManager()",
					}).Infof("distributeMsg failed.")

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
	klog.Info("Dmanager Received message: %v", m)
	msg, ok := m.(model.Message)
	if !ok {
		klog.Errorf("Distribute message, msg is nil")
		return errors.New("Distribute message, msg is nil")
	}
	message := dmtype.DMMessage{Msg: &msg}
	if message.Msg.GetParentID() != "" {
		confirmMsg := dmtype.DMMessage{Msg: model.NewMessage(message.Msg.GetParentID()), Action: dmcommon.Confirm}
		if err := dm.DMContexts.CommTo(dmcommon.CommModule, &confirmMsg); err != nil {
			return err
		}
	}

	if !classifyMsg(&message) {
		return nil
	}
	if ActionModuleMap == nil {
		initActionModuleMap()
	}

	if moduleName, exist := ActionModuleMap[message.Action]; exist {
		//how to deal write channel error
		klog.Infof("Send msg to the %s module in twin", moduleName)
		if err := dm.DMContexts.CommTo(moduleName, &message); err != nil {
			logrus.WithFields(logrus.Fields{
				"module":     "dmanager",
				"moduleName": moduleName,
				"message":    message,
				"func":       "distributeMsg()",
			}).Errorf("Error: dm.DMContexts.CommTo(moduleName, &message)")
			return err
		}
	} else {
		klog.Info("Not found deal module for msg")
		return errors.New("Not found deal module for msg")
	}

	return nil
}

func classifyMsg(message *dmtype.DMMessage) bool {
	if EventActionMap == nil {
		initEventActionMap()
	}

	var identity string
	var action string
	msgSource := message.Msg.GetSource()
	if strings.Compare(msgSource, "edgemgr") == 0 {
		klog.Infof("Edgemgr msg is %v", message.Msg)
		idLoc := 3
		topic := message.Msg.GetResource()
		topicByte, err := base64.URLEncoding.DecodeString(topic)
		if err != nil {
			return false
		}
		topic = string(topicByte)

		klog.Infof("classify the msg with the topic %s", topic)
		splitString := strings.Split(topic, "/")
		if len(splitString) == 4 {
			// Events to Actions
			if strings.HasPrefix(topic, dmcommon.LifeCycleConnectETPrefix) {
				action = dmcommon.LifeCycle
			} else if strings.HasPrefix(topic, dmcommon.LifeCycleDisconnectETPrefix) {
				action = dmcommon.LifeCycle
			} else {
				return false
			}
		} else {
			//identity refers to "dht11-sensor-1" as "$hw/events/device/dht11-sensor-1/twin/get/result"
			identity = splitString[idLoc]
			loc := strings.Index(topic, identity)
			nextLoc := loc + len(identity)
			// prefix refers to "$hw/events/device/" here
			prefix := topic[0:loc]
			// suffix refers to "/twin/get/result" here
			suffix := topic[nextLoc:]
			klog.Infof("ClassfyMsg: Identity: %s, prefix: %s, suffix: %s", identity, prefix, suffix)
			//klog.Infof("%s %s", prefix, suffix)
			if v, exist := EventActionMap[prefix][suffix]; exist {
				action = v
			} else {
				return false
			}
		}

		message.Msg.Content = []byte((message.Msg.Content).(string))
		message.Identity = identity
		message.Action = action
		klog.Infof("Classify the msg to action %s", action)
		return true

	} else if (strings.Compare(msgSource, modules.DMgrGroup) == 0) || (strings.Compare(msgSource, "devicecontroller") == 0) {
		// Here to handle real device msgs
		klog.Infof("Msg source2 is %s, Msg is %s", msgSource, message.Msg)
		logrus.WithFields(logrus.Fields{
			"module": "dmanager",
			"func":   "classifyMsg()",
			"source": msgSource,
			"msg":    message.Msg,
		}).Infof("Classify Msg!")
		switch message.Msg.Content.(type) {
		case []byte:
			klog.Info("Message content type is []byte, no need to marshal again")
		default:
			content, err := json.Marshal(message.Msg.Content)
			if err != nil {
				return false
			}
			message.Msg.Content = content
		}
		if strings.Contains(message.Msg.Router.Resource, "membership/detail") {
			message.Action = dmcommon.MemDetailResult
			return true
		} else if strings.Contains(message.Msg.Router.Resource, "membership") {
			message.Action = dmcommon.MemUpdated
			return true
		} else if strings.Contains(message.Msg.Router.Resource, "twin/cloud_updated") {
			message.Action = dmcommon.TwinCloudSync
			resources := strings.Split(message.Msg.Router.Resource, "/")
			message.Identity = resources[1]
			return true
		} else if strings.Contains(message.Msg.Router.Operation, "updated") {
			resources := strings.Split(message.Msg.Router.Resource, "/")
			if len(resources) == 2 && strings.Compare(resources[0], "device") == 0 {
				message.Action = dmcommon.DeviceUpdated
				message.Identity = resources[1]
			}
			return true
		}
		return false
	} else if strings.Compare(msgSource, "edgehub") == 0 {
		klog.Infof("Msg from edgehub is %s", message.Msg)
		if strings.Compare(message.Msg.Router.Resource, "node/connection") == 0 {
			message.Action = dmcommon.LifeCycle
			return true
		}
		return false
	}
	return false
}

func initEventActionMap() {
	EventActionMap = make(map[string]map[string]string)
	EventActionMap[dmcommon.MemETPrefix] = make(map[string]string)
	EventActionMap[dmcommon.DeviceETPrefix] = make(map[string]string)

	EventActionMap[dmcommon.MemETPrefix][dmcommon.MemETDetailResultSuffix] = dmcommon.MemDetailResult
	EventActionMap[dmcommon.MemETPrefix][dmcommon.MemETUpdateSuffix] = dmcommon.MemUpdated
	EventActionMap[dmcommon.MemETPrefix][dmcommon.MemETGetSuffix] = dmcommon.MemGet

	EventActionMap[dmcommon.DeviceETPrefix][dmcommon.DeviceETStateGetSuffix] = dmcommon.DeviceStateGet
	EventActionMap[dmcommon.DeviceETPrefix][dmcommon.DeviceETUpdatedSuffix] = dmcommon.DeviceUpdated
	EventActionMap[dmcommon.DeviceETPrefix][dmcommon.DeviceETStateUpdateSuffix] = dmcommon.DeviceStateUpdate

	EventActionMap[dmcommon.DeviceETPrefix][dmcommon.TwinETUpdateSuffix] = dmcommon.TwinUpdate
	EventActionMap[dmcommon.DeviceETPrefix][dmcommon.TwinETCloudSyncSuffix] = dmcommon.TwinCloudSync
	EventActionMap[dmcommon.DeviceETPrefix][dmcommon.TwinETGetSuffix] = dmcommon.TwinGet
}

func initActionModuleMap() {
	ActionModuleMap = make(map[string]string)
	//membership twin device event , not lifecycle event
	ActionModuleMap[dmcommon.MemDetailResult] = dmcommon.MemModule
	ActionModuleMap[dmcommon.MemGet] = dmcommon.MemModule
	ActionModuleMap[dmcommon.MemUpdated] = dmcommon.MemModule

	// Twin updated handle transfer to MemModule
	ActionModuleMap[dmcommon.TwinGet] = dmcommon.MemModule
	ActionModuleMap[dmcommon.TwinUpdate] = dmcommon.MemModule
	ActionModuleMap[dmcommon.TwinCloudSync] = dmcommon.MemModule

	ActionModuleMap[dmcommon.DeviceUpdated] = dmcommon.DeviceModule
	ActionModuleMap[dmcommon.DeviceStateGet] = dmcommon.DeviceModule
	ActionModuleMap[dmcommon.DeviceStateUpdate] = dmcommon.DeviceModule
	ActionModuleMap[dmcommon.Connected] = dmcommon.CommModule
	ActionModuleMap[dmcommon.Disconnected] = dmcommon.CommModule
	ActionModuleMap[dmcommon.LifeCycle] = dmcommon.CommModule
	ActionModuleMap[dmcommon.Confirm] = dmcommon.CommModule
}
