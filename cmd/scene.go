package cmd

import (
	"encoding/binary"

	mesh "github.com/AJGherardi/GoMeshController"
	"github.com/AJGherardi/HomeHub/model"
)

// SceneStore adds a scene to a group and sends a SceneStore message
func SceneStore(store *model.Store, controller mesh.Controller, groupAddr uint16, name string) uint16 {
	group := store.Groups[groupAddr]
	netData := store.NetData
	// Get and increment next scene number
	sceneNumber := netData.GetNextSceneNumber()
	netData.IncrementNextSceneNumber()
	// Store scene
	group.AddScene(name, sceneNumber)
	controller.SendStoreMessage(sceneNumber, groupAddr, group.KeyIndex)
	return sceneNumber
}

// SceneRecall sends a recall message to devices in the group
func SceneRecall(store *model.Store, controller mesh.Controller, groupAddr, sceneNumber uint16) {
	group := store.Groups[groupAddr]
	controller.SendRecallMessage(sceneNumber, groupAddr, group.KeyIndex)
}

// SceneDelete sends a delete message to devices in the group and removes the scene from the group
func SceneDelete(store *model.Store, controller mesh.Controller, groupAddr, sceneNumber uint16) {
	group := store.Groups[groupAddr]
	group.DeleteScene(sceneNumber)
	controller.SendDeleteMessage(sceneNumber, groupAddr, group.KeyIndex)
}

// EventBind sends a event bind msg to the element and sets the elements state to the scene number
func EventBind(store *model.Store, controller mesh.Controller, groupAddr, devAddr, elemAddr, sceneNumber uint16) {
	group := store.Groups[groupAddr]
	device := group.Devices[devAddr]
	// Convert sceneNumber to bytes for device state struct
	var sceneNumberBytes []byte
	binary.BigEndian.PutUint16(sceneNumberBytes, sceneNumber)
	device.UpdateState(elemAddr, sceneNumberBytes)
	controller.SendBindMessage(sceneNumber, elemAddr, group.KeyIndex)
}
