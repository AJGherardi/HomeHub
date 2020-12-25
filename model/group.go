package model

import "errors"

// MakeGroup makes a new group with the given addr
func MakeGroup(name string, addr, keyIndex uint16) Group {
	group := Group{
		Name:     name,
		KeyIndex: keyIndex,
		Devices:  map[uint16]*Device{},
		Scenes:   map[uint16]*Scene{},
	}
	return group
}

// Group holds a collection of devices and its app key
type Group struct {
	Name     string
	KeyIndex uint16
	Devices  map[uint16]*Device
	Scenes   map[uint16]*Scene
}

// GetDevice gets the ref to device with given addr
func (g *Group) GetDevice(addr uint16) (*Device, error) {
	device := g.Devices[addr]
	if device == nil {
		return nil, errors.New("No Device with given address")
	}
	return device, nil
}

// AddDevice adds a device to the group
func (g *Group) AddDevice(addr uint16, device Device) {
	g.Devices[addr] = &device
}

// RemoveDevice removes the device from the group
func (g *Group) RemoveDevice(addr uint16) {
	delete(g.Devices, addr)
}

// GetScene gets the ref to scene with given number
func (g *Group) GetScene(number uint16) (*Scene, error) {
	scene := g.Scenes[number]
	if scene == nil {
		return nil, errors.New("No Scene with given number")
	}
	return scene, nil
}

// AddScene adds a scene to a the group
func (g *Group) AddScene(name string, number uint16) {
	g.Scenes[number] = &Scene{Name: name}

}

// RemoveScene removes a scene from the group
func (g *Group) RemoveScene(number uint16) {
	delete(g.Scenes, number)
}

// Scene holds the scenes name and number
type Scene struct {
	Name string
}
