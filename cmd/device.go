package cmd

import (
	"errors"
	"reflect"
	"time"

	mesh "github.com/AJGherardi/GoMeshController"
	"github.com/AJGherardi/HomeHub/model"
)

// AddDevice provisions configures and adds the device at the given uuid
func AddDevice(store *model.Store, controller mesh.Controller, name string, uuid []byte, groupAddr uint16, nodeAdded chan uint16) (uint16, error) {
	// Get group
	group, getErr := store.GetGroup(groupAddr)
	if getErr != nil {
		return 0, getErr
	}
	// Provision device
	if err := controller.Provision(uuid); err != nil {
		return 0, errors.New("Device Setup Failed")
	}
	// Wait for node added
	var nodeAddr uint16
	select {
	case addr := <-nodeAdded:
		nodeAddr = addr
	case <-time.After(10 * time.Second):
		// Timeout after 10 seconds
		return 0, errors.New("Timeout adding device")
	}
	// Make device object
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
	// Configure Device
	sendErr := controller.ConfigureNode(nodeAddr, group.KeyIndex)
	time.Sleep(100 * time.Millisecond)
	// If device is a 2 plug outlet
	if reflect.DeepEqual(uuid[6:8], []byte{0x00, 0x02}) {
		// Set type and add elements
		elemAddr0 := device.AddElem(name+"-0", "onoff", nodeAddr)
		sendErr = controller.ConfigureElem(groupAddr, nodeAddr, elemAddr0, group.KeyIndex)
		time.Sleep(100 * time.Millisecond)
		elemAddr1 := device.AddElem(name+"-1", "onoff", nodeAddr)
		sendErr = controller.ConfigureElem(groupAddr, nodeAddr, elemAddr1, group.KeyIndex)
	}
	// If device is a button
	if reflect.DeepEqual(uuid[6:8], []byte{0x00, 0x01}) {
		// Set type and add elements
		elemAddr0 := device.AddElem(name+"-0", "event", nodeAddr)
		sendErr = controller.ConfigureElem(groupAddr, nodeAddr, elemAddr0, group.KeyIndex)
		time.Sleep(1000 * time.Millisecond)
		elemAddr1 := device.AddElem(name+"-1", "event", nodeAddr)
		sendErr = controller.ConfigureElem(groupAddr, nodeAddr, elemAddr1, group.KeyIndex)
		time.Sleep(3000 * time.Millisecond)
		elemAddr2 := device.AddElem(name+"-2", "event", nodeAddr)
		sendErr = controller.ConfigureElem(groupAddr, nodeAddr, elemAddr2, group.KeyIndex)
		time.Sleep(3000 * time.Millisecond)
		elemAddr3 := device.AddElem(name+"-3", "event", nodeAddr)
		sendErr = controller.ConfigureElem(groupAddr, nodeAddr, elemAddr3, group.KeyIndex)
	}
	// If device is not setup do not procead
	if sendErr != nil {
		return 0, errors.New("Device Setup Failed")
	}
	// Add device to group
	group.AddDevice(nodeAddr, device)
	return nodeAddr, nil
}

// RemoveDevice sends are reset message and removes the device from the group
func RemoveDevice(store *model.Store, controller mesh.Controller, groupAddr, devAddr uint16) error {
	// Get group
	group, getErr := store.GetGroup(groupAddr)
	if getErr != nil {
		return getErr
	}
	// Send reset payload
	sendErr := controller.ResetNode(devAddr)
	if sendErr != nil {
		return errors.New("Failed to remove device")
	}
	// Remove device from group
	group.RemoveDevice(devAddr)
	return nil
}

// SetState sends a state message to the element at the given address
func SetState(store *model.Store, controller mesh.Controller, state []byte, groupAddr, elemAddr uint16) error {
	// Get group
	group, getErr := store.GetGroup(groupAddr)
	if getErr != nil {
		return getErr
	}
	// Send State
	if true {
		// Send msg
		sendErr := controller.SendMessage(
			state[0],
			elemAddr,
			group.KeyIndex,
		)
		if sendErr != nil {
			return errors.New("Failed to send message")
		}
	}
	return nil
}

// ReadState Gets the state from a elem on a device
func ReadState(store *model.Store, groupAddr, devAddr, elemAddr uint16) ([]byte, error) {
	// Get group
	group, getErr := store.GetGroup(groupAddr)
	if getErr != nil {
		return []byte{}, getErr
	}
	// Get device
	device, getErr := group.GetDevice(devAddr)
	if getErr != nil {
		return []byte{}, getErr
	}
	state, getErr := device.GetState(elemAddr)
	return state, getErr
}

// UpdateState Sets the state of a elem on the device
func UpdateState(store *model.Store, elemAddr uint16, state byte) error {
	// Get refrence to device with element
	for _, group := range store.Groups {
		for _, device := range group.Devices {
			for addr := range device.Elements {
				if addr == elemAddr {
					// Update element state
					setErr := device.UpdateState(elemAddr, []byte{state})
					return setErr

				}
			}
		}
	}
	return errors.New("Element not found")
}
