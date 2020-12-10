package cmd

import (
	"reflect"
	"time"

	mesh "github.com/AJGherardi/GoMeshController"
	"github.com/AJGherardi/HomeHub/model"
)

// AddDevice provisions configures and adds the device at the given uuid
func AddDevice(store *model.Store, controller mesh.Controller, name string, uuid []byte, groupAddr uint16, nodeAdded chan uint16) uint16 {
	// Provision device
	controller.Provision(uuid)
	// Wait for node added
	nodeAddr := <-nodeAdded
	device := model.Device{}
	// If device is a 2 plug outlet
	if reflect.DeepEqual(uuid[6:8], []byte{0x00, 0x02}) {
		device = model.MakeDevice(
			"2Outlet",
			nodeAddr,
		)
	}
	if reflect.DeepEqual(uuid[6:8], []byte{0x00, 0x01}) {
		device = model.MakeDevice(
			"4Button",
			nodeAddr,
		)
	}
	// Get group
	group := store.Groups[groupAddr]
	// Configure Device
	controller.ConfigureNode(nodeAddr, group.KeyIndex)
	time.Sleep(100 * time.Millisecond)
	// If device is a 2 plug outlet
	if reflect.DeepEqual(uuid[6:8], []byte{0x00, 0x02}) {
		// Set type and add elements
		elemAddr0 := device.AddElem(name+"-0", "onoff", nodeAddr)
		controller.ConfigureElem(groupAddr, nodeAddr, elemAddr0, group.KeyIndex)
		time.Sleep(100 * time.Millisecond)
		elemAddr1 := device.AddElem(name+"-1", "onoff", nodeAddr)
		controller.ConfigureElem(groupAddr, nodeAddr, elemAddr1, group.KeyIndex)
	}
	// If device is a button
	if reflect.DeepEqual(uuid[6:8], []byte{0x00, 0x01}) {
		// Set type and add elements
		elemAddr0 := device.AddElem(name+"-0", "event", nodeAddr)
		controller.ConfigureElem(groupAddr, nodeAddr, elemAddr0, group.KeyIndex)
		time.Sleep(1000 * time.Millisecond)
		elemAddr1 := device.AddElem(name+"-1", "event", nodeAddr)
		controller.ConfigureElem(groupAddr, nodeAddr, elemAddr1, group.KeyIndex)
		time.Sleep(3000 * time.Millisecond)
		elemAddr2 := device.AddElem(name+"-2", "event", nodeAddr)
		controller.ConfigureElem(groupAddr, nodeAddr, elemAddr2, group.KeyIndex)
		time.Sleep(3000 * time.Millisecond)
		elemAddr3 := device.AddElem(name+"-3", "event", nodeAddr)
		controller.ConfigureElem(groupAddr, nodeAddr, elemAddr3, group.KeyIndex)
	}
	// Add device to group
	group.AddDevice(nodeAddr, device)
	return nodeAddr
}

// RemoveDevice sends are reset message and removes the device from the group
func RemoveDevice(store *model.Store, controller mesh.Controller, groupAddr, addr uint16) {
	// Get device
	group := store.Groups[groupAddr]
	// Send reset payload
	controller.ResetNode(addr)
	// Remove device from group
	group.RemoveDevice(addr)
}

// SetState sends a state message to the element at the given address
func SetState(store *model.Store, controller mesh.Controller, state []byte, groupAddr, addr uint16) {
	// Get appKey from group
	group := store.Groups[groupAddr]
	// Send State
	if true {
		// Send msg
		controller.SendMessage(
			state[0],
			addr,
			group.KeyIndex,
		)
	}
}

// ReadStates Gets the state from a elem on a device
func ReadStates(store *model.Store, groupAddr, devAddr, elemAddr uint16) []byte {
	device := store.Groups[groupAddr].Devices[devAddr]
	state := device.GetState(elemAddr)
	return state
}

// UpdateState Sets the state of a elem on the device
func UpdateState(store *model.Store, addr uint16, state byte) {
	// Get refrence to device with element
	for _, group := range store.Groups {
		for _, device := range group.Devices {
			for elemAddr := range device.Elements {
				if elemAddr == addr {
					// Update element state
					device.UpdateState(addr, []byte{state})
				}
			}

		}
	}
}
