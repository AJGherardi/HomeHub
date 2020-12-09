package model

// MakeDevice makes a new device with the given addr
func MakeDevice(deviceType string, addr uint16) Device {
	return Device{
		Type:     deviceType,
		Elements: map[uint16]*Element{},
	}
}

// Device holds the name addr and type of device
type Device struct {
	Type     string
	Elements map[uint16]*Element
}

// AddElem adds a element to the device
func (d *Device) AddElem(name, stateType string, devAddr uint16) uint16 {
	// Check if first elem
	if len(d.Elements) == 0 {
		// Create element with device main addr
		element := Element{
			Name:      name,
			State:     []byte{0x00},
			StateType: stateType,
		}
		d.Elements[devAddr] = &element
		return devAddr
	}
	// If not create element using incremented address
	addr := devAddr + uint16(len(d.Elements))
	element := Element{
		Name:      name,
		State:     []byte{0x00},
		StateType: stateType,
	}
	d.Elements[addr] = &element
	return addr
}

// UpdateState updates the state of the element with the given address
func (d *Device) UpdateState(addr uint16, state []byte) {
	d.Elements[addr].State = state
}

// GetState returns the state of a element
func (d *Device) GetState(addr uint16) []byte {
	return d.Elements[addr].State
}

// Element holds an elements name, addr and its state
type Element struct {
	Name      string
	State     []byte
	StateType string
}
