package cmd

import (
	"encoding/binary"
	"errors"

	mesh "github.com/AJGherardi/GoMeshController"
	"github.com/AJGherardi/HomeHub/model"
)

// SceneStore adds a scene to a group and sends a SceneStore message
func SceneStore(store *model.Store, controller mesh.Controller, groupAddr uint16, name string) (uint16, error) {
	group := store.Groups[groupAddr]
	group, getErr := store.GetGroup(groupAddr)
	if getErr != nil {
		return 0, getErr
	}
	netData, getErr := store.GetNetData()
	if getErr != nil {
		return 0, getErr
	}
	// Get and increment next scene number
	sceneNumber := netData.GetNextSceneNumber()
	netData.IncrementNextSceneNumber()
	sendErr := controller.SendStoreMessage(sceneNumber, groupAddr, group.KeyIndex)
	// Send scene store message
	if sendErr != nil {
		return 0, errors.New("Failed to send scene store message")
	}
	// Store scene
	group.AddScene(name, sceneNumber)
	return sceneNumber, nil
}

// SceneRecall sends a recall message to devices in the group
func SceneRecall(store *model.Store, controller mesh.Controller, groupAddr, sceneNumber uint16) error {
	group, getErr := store.GetGroup(groupAddr)
	if getErr != nil {
		return getErr
	}
	sendErr := controller.SendRecallMessage(sceneNumber, groupAddr, group.KeyIndex)
	if sendErr != nil {
		return errors.New("Failed to send scene recall message")
	}
	return nil
}

// SceneDelete sends a delete message to devices in the group and removes the scene from the group
func SceneDelete(store *model.Store, controller mesh.Controller, groupAddr, sceneNumber uint16) error {
	group, getErr := store.GetGroup(groupAddr)
	if getErr != nil {
		return getErr
	}
	sendErr := controller.SendDeleteMessage(sceneNumber, groupAddr, group.KeyIndex)
	if sendErr != nil {
		return errors.New("Failed to send scene delete message")
	}
	group.DeleteScene(sceneNumber)
	return nil
}

// EventBind sends a event bind msg to the element and sets the elements state to the scene number
func EventBind(store *model.Store, controller mesh.Controller, groupAddr, devAddr, elemAddr, sceneNumber uint16) error {
	group, getErr := store.GetGroup(groupAddr)
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
	sendErr := controller.SendBindMessage(sceneNumber, elemAddr, group.KeyIndex)
	if sendErr != nil {
		return errors.New("Failed to send event bind message")
	}
	device.UpdateState(elemAddr, sceneNumberBytes)
	return nil
}
