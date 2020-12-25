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
	n.Store.RemoveGroup(groupAddr)
	return sendErr
}

// AddGroup creates a group in the store with the given name
func (n *Network) AddGroup(name string) (uint16, error) {
	if n.Store.GetConfigured() == false {
		return 0, errors.New("Not configured")
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

// GetGroups returns a map of refreences to all of the groups in the network
func (n *Network) GetGroups() map[uint16]*model.Group {
	return n.Store.Groups
}

// GetGroup returns a refrence to the group with the given address
func (n *Network) GetGroup(groupAddr uint16) (*model.Group, error) {
	group, err := n.Store.GetGroup(groupAddr)
	return group, err
}
