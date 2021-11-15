package cdmtype

import "github.com/kubeedge/beehive/pkg/core/model"

//DMMessage the struct of message for commutinating between cloud and edge
type DMMessage struct {
	Msg      *model.Message
	Identity string
	Action   string
	Type     string
}
