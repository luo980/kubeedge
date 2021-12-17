package cdmtype

import (
	"github.com/kubeedge/beehive/pkg/core/model"
)

//CDMMessage the struct of message for communicating between cloud and edge
type CDMMessage struct {
	Msg      *model.Message
	Identity string
	Action   string
	Type     string
}

type ReqMMessage struct {
	Identity string
	Action   string
	Type     string
	Body	 []byte
	RawQuery	string
}
