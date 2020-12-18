package cmd

import (
	mesh "github.com/AJGherardi/GoMeshController"
	"github.com/AJGherardi/HomeHub/model"
)

// RemoveGroup deletes the group from the store and sends reset messages to all the devices in the group
func RemoveGroup(store *model.Store, controller mesh.Controller, groupAddr uint16) error {
	// Get groupAddr
	group, getErr := store.GetGroup(groupAddr)
	if getErr != nil {
		return getErr
	}
	// Reset devices
	var sendErr error
	for addr := range group.Devices {
		err := controller.ResetNode(addr)
		if err != nil {
			sendErr = err
		}
	}
	// Remove the group
	delete(store.Groups, groupAddr)
	return sendErr
}

// AddGroup creates a group in the store with the given name
func AddGroup(store *model.Store, name string) (uint16, error) {
	netData, getErr := store.GetNetData()
	if getErr != nil {
		return 0, getErr
	}
	// Get net values
	groupAddr := netData.GetNextGroupAddr()
	// Add a group
	group := model.MakeGroup(name, groupAddr, 0x0000)
	// Add group to store
	store.AddGroup(groupAddr, group)
	// Update net data
	netData.IncrementNextGroupAddr()
	return groupAddr, nil
}
