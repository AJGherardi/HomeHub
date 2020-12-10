package cmd

import (
	mesh "github.com/AJGherardi/GoMeshController"
	"github.com/AJGherardi/HomeHub/model"
)

// RemoveGroup deletes the group from the store and sends reset messages to all the devices in the group
func RemoveGroup(store *model.Store, controller mesh.Controller, addr uint16) {
	// Get groupAddr
	group := store.Groups[addr]
	// Reset devices
	for addr := range group.Devices {
		controller.ResetNode(addr)
	}
	// Remove the group
	delete(store.Groups, addr)
}

// AddGroup creates a group in the store with the given name
func AddGroup(store *model.Store, name string) uint16 {
	netData := store.NetData
	// Get net values
	groupAddr := netData.GetNextGroupAddr()
	// Add a group
	group := model.MakeGroup(name, groupAddr, 0x0000)
	store.Groups[groupAddr] = &group
	// Update net data
	netData.IncrementNextGroupAddr()
	return groupAddr
}
