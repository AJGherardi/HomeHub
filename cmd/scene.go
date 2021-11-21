package cmd

import (
	"encoding/binary"
	"errors"
)

// SceneStore adds a scene to a group and sends a SceneStore message
func (n *Network) SceneStore(groupAddr uint16, name string) (uint16, error) {
	if n.Store.GetConfigured() == false {
		return 0, errors.New("Not Configured")
	}
	group, getErr := n.Store.GetGroup(groupAddr)
	if getErr != nil {
		return 0, getErr
	}
	// Get and increment next scene number
	sceneNumber := n.Store.GetNextSceneNumber()
	n.Store.IncrementNextSceneNumber()
	sendErr := n.Controller.SendStoreMessage(sceneNumber, groupAddr, group.KeyIndex)
	// Send scene store message
	if sendErr != nil {
		return 0, errors.New("Failed to send scene store message")
	}
	// Store scene
	group.AddScene(name, sceneNumber)
	return sceneNumber, nil
}

// SceneRecall sends a recall message to devices in the group
func (n *Network) SceneRecall(groupAddr, sceneNumber uint16) error {
	group, getErr := n.Store.GetGroup(groupAddr)
	if getErr != nil {
		return getErr
	}
	sendErr := n.Controller.SendRecallMessage(sceneNumber, groupAddr, group.KeyIndex)
	if sendErr != nil {
		return errors.New("Failed to send scene recall message")
	}
	return nil
}

// SceneDelete sends a delete message to devices in the group and removes the scene from the group
func (n *Network) SceneDelete(groupAddr, sceneNumber uint16) error {
	group, getErr := n.Store.GetGroup(groupAddr)
	if getErr != nil {
		return getErr
	}
	sendErr := n.Controller.SendDeleteMessage(sceneNumber, groupAddr, group.KeyIndex)
	if sendErr != nil {
		return errors.New("Failed to send scene delete message")
	}
	group.RemoveScene(sceneNumber)
	return nil
}

// EventBind sends a event bind msg to the element and sets the elements state to the scene number
func (n *Network) EventBind(groupAddr, devAddr, elemAddr, sceneNumber uint16) error {
	group, getErr := n.Store.GetGroup(groupAddr)
	if getErr != nil {
		return getErr
	}
	device, getErr := group.GetDevice(devAddr)
	if getErr != nil {
		return getErr
	}
	// Convert sceneNumber to bytes for device state struct
	var sceneNumberBytes []byte
	binary.BigEndian.PutUint16(sceneNumberBytes, sceneNumber)
	sendErr := n.Controller.SendBindMessage(sceneNumber, elemAddr, group.KeyIndex)
	if sendErr != nil {
		return errors.New("Failed to send event bind message")
	}
	device.UpdateState(elemAddr, sceneNumberBytes)
	return nil
}
