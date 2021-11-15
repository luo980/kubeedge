package receiver

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kubeedge/beehive/pkg/core/model"
	receiverconstants "github.com/kubeedge/kubeedge/cloud/pkg/cdmanager/constants"
	constants "github.com/kubeedge/kubeedge/common/constants"
)

// BuildResource return a string as "beehive/pkg/core/model".Message.Router.Resource
func BuildResource(nodeID, resourceType, resourceID string) (resource string, err error) {
	if nodeID == "" || resourceType == "" {
		err = fmt.Errorf("required parameter are not set (node id, namespace or resource type)")
		return
	}
	resource = fmt.Sprintf("%s%s%s%s%s", receiverconstants.ResourceNode, constants.ResourceSep, nodeID, constants.ResourceSep, resourceType)
	if resourceID != "" {
		resource += fmt.Sprintf("%s%s", constants.ResourceSep, resourceID)
	}
	return
}

// GetDeviceID returns the ID of the device
func GetDeviceID(resource string) (string, error) {
	res := strings.Split(resource, "/")
	if len(res) >= receiverconstants.ResourceDeviceIDIndex+1 && res[receiverconstants.ResourceDeviceIndex] == receiverconstants.ResourceDevice {
		return res[receiverconstants.ResourceDeviceIDIndex], nil
	}
	return "", errors.New("failed to get device id")
}

// GetResourceType returns the resourceType of message received from edge
func GetResourceType(resource string) (string, error) {
	if strings.Contains(resource, receiverconstants.ResourceTypeTwinEdgeUpdated) {
		return receiverconstants.ResourceTypeTwinEdgeUpdated, nil
	} else if strings.Contains(resource, receiverconstants.ResourceTypeMembershipDetail) {
		return receiverconstants.ResourceTypeMembershipDetail, nil
	}

	return "", fmt.Errorf("unknown resource, found: %s", resource)
}

// GetNodeID from "beehive/pkg/core/model".Message.Router.Resource
func GetNodeID(msg model.Message) (string, error) {
	sli := strings.Split(msg.GetResource(), constants.ResourceSep)
	if len(sli) <= receiverconstants.ResourceNodeIDIndex {
		return "", fmt.Errorf("node id not found")
	}
	return sli[receiverconstants.ResourceNodeIDIndex], nil
}
