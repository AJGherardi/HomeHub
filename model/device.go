package model

import (
	"reflect"

	"github.com/AJGherardi/HomeHub/utils"
)

// MakeDevice makes a new device with the given addr
func MakeDevice(deviceType string, addr []byte, db DB) Device {
	device := Device{
		Type: deviceType,
		Addr: addr,
	}
	db.InsertDevice(device)
	return device
}

// Device holds the name addr and type of device
type Device struct {
	Type     string
	Addr     []byte
	Elements []Element
}

// AddElem adds a element to the device
func (d *Device) AddElem(name, stateType string, db DB) []byte {
	// Check if first elem
	if len(d.Elements) == 0 {
		// Create element with device main addr
		element := Element{
			Name: name,
			Addr: d.Addr,
			State: State{
				State:     []byte{0x00},
				StateType: stateType,
			},
		}
		d.Elements = append(d.Elements, element)
		return d.Addr
	}
	// If not create element using incremented address
	addr := utils.IncrementAddr(d.Elements[len(d.Elements)-1].Addr)
	element := Element{
		Name: name,
		Addr: addr,
		State: State{
			State:     []byte{0x00},
			StateType: stateType,
		},
	}
	d.Elements = append(d.Elements, element)
	db.UpdateDevice(*d)
	return addr
}

// UpdateState updates the state of a element
func (d *Device) UpdateState(index int, state []byte, db DB) {
	d.Elements[index].State.State = state
	db.UpdateDevice(*d)
}

// UpdateStateUsingAddr updates the state of the element with the given address
func (d *Device) UpdateStateUsingAddr(addr, state []byte, db DB) {
	for i, element := range d.Elements {
		if reflect.DeepEqual(element.Addr, addr) {
			d.Elements[i].State.State = state
			db.UpdateDevice(*d)
		}
	}
}

// GetState returns the state of a element
func (d *Device) GetState(index int) State {
	return d.Elements[index].State
}

// GetElemAddr returns the address of a element
func (d *Device) GetElemAddr(index int) []byte {
	return d.Elements[index].Addr
}

// Element holds an elements name, addr and its state
type Element struct {
	Name  string
	Addr  []byte
	State State
}

// State has a value and a type
type State struct {
	State     []byte
	StateType string
}
