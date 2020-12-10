package graph

import (
	"github.com/AJGherardi/HomeHub/generated"
	"github.com/AJGherardi/HomeHub/model"
)

func toElementResponseSlice(elements map[uint16]*model.Element) []*generated.ElementResponse {
	elementResponses := []*generated.ElementResponse{}
	for i := range elements {
		elementResponses = append(
			elementResponses,
			toElementResponse(i, elements[i]),
		)
	}
	return elementResponses
}

func toDeviceResponseSlice(devices map[uint16]*model.Device) []*generated.DeviceResponse {
	deviceResponses := []*generated.DeviceResponse{}
	for i := range devices {
		deviceResponses = append(
			deviceResponses,
			toDeviceResponse(i, devices[i]),
		)
	}
	return deviceResponses
}

func toSceneResponseSlice(scenes map[uint16]*model.Scene) []*generated.SceneResponse {
	sceneResponses := []*generated.SceneResponse{}
	for i := range scenes {
		sceneResponses = append(
			sceneResponses,
			toSceneResponse(i, scenes[i]),
		)
	}
	return sceneResponses
}

func toElementResponse(addr uint16, element *model.Element) *generated.ElementResponse {
	return &generated.ElementResponse{Addr: int(addr), Element: element}
}

func toDeviceResponse(addr uint16, device *model.Device) *generated.DeviceResponse {
	return &generated.DeviceResponse{Addr: int(addr), Device: device}
}

func toSceneResponse(number uint16, scene *model.Scene) *generated.SceneResponse {
	return &generated.SceneResponse{Number: int(number), Scene: scene}
}
