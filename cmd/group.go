package cmd

import (
	"errors"

	"github.com/AJGherardi/HomeHub/model"
)

// RemoveGroup deletes the group from the store and sends reset messages to all the devices in the group
func (n *Network) RemoveGroup(groupAddr uint16) error {
	// Get groupAddr
	group, getErr := n.Store.GetGroup(groupAddr)
	if getErr != nil {
		return getErr
	}
	// Reset devices
	var sendErr error
	for addr := range group.Devices {
		err := n.Controller.ResetNode(addr)
		if err != nil {
			sendErr = err
		}
	}
	// Remove the group
	delete(n.Store.Groups, groupAddr)
	return sendErr
}

// AddGroup creates a group in the store with the given name
func (n *Network) AddGroup(name string) (uint16, error) {
	if n.Store.GetConfigured() != false {
		return 0, errors.New("Not Configured")
	}
	// Get net values
	groupAddr := n.Store.GetNextGroupAddr()
	// Add a group
	group := model.MakeGroup(name, groupAddr, 0x0000)
	// Add group to store
	n.Store.AddGroup(groupAddr, group)
	// Update net data
	n.Store.IncrementNextGroupAddr()
	return groupAddr, nil
}
