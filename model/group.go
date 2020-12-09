package model

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

// AddDevice adds a device to the group
func (g *Group) AddDevice(addr uint16, device Device) {
	g.Devices[addr] = &device
}

// RemoveDevice removes the device from the group
func (g *Group) RemoveDevice(addr uint16) {
	delete(g.Devices, addr)
}

// AddScene adds a scene to a the group
func (g *Group) AddScene(name string, number uint16) {
	g.Scenes[number] = &Scene{Name: name}

}

// DeleteScene removes a scene from the group
func (g *Group) DeleteScene(number uint16) {
	delete(g.Scenes, number)
}

// Scene holds the scenes name and number
type Scene struct {
	Name string
}
