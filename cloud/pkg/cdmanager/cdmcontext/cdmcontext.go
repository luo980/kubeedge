package cdmcontext

import (
	"errors"
	"github.com/kubeedge/beehive/pkg/core/model"
	"github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/cdmtype"
	"github.com/kubeedge/kubeedge/cloud/pkg/common/modules"
)

type CDMContext struct {
	GroupID			string
	CommChan		map[string]chan interface{}
	ConfirmChan		chan interface{}
}

func InitCDMContext() (*CDMContext, error){
	return &CDMContext{
		GroupID:     "",
		CommChan:    make(map[string]chan interface{}),
		ConfirmChan: make(chan interface{}, 1000),
	}, nil
}

func (cdmc *CDMContext) CommTo(cdmcName string, content interface{}) error{
	if v, exist := cdmc.CommChan[cdmcName]; exist{
		v <- content
		return nil
	}
	return errors.New("not found chan to communicate")
}

func (cdmc *CDMContext) Send(identity string, action string, module string, msg *model.Message) error {
	cdmMsg := &cdmtype.CDMMessage{
		Action:   action,
		Identity: identity,
		Type:     module,
		Msg:      msg}
	return cdmc.CommTo(module, cdmMsg)
}

func (cdmc *CDMContext) BuildModelMessage(group string, parentID string, resource string, operation string, content interface{}) *model.Message {
	msg := model.NewMessage(parentID)
	msg.BuildRouter(modules.CDmanagerGroupName, group, resource, operation)
	msg.Content = content
	return msg
}