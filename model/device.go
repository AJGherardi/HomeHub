package model

import (
	"reflect"

	"github.com/AJGherardi/HomeHub/utils"
)

func MakeDevice(name, deviceType string, addr []byte, db DB) Device {
	device := Device{
		Name: name,
		Type: deviceType,
		Addr: addr,
	}
	db.InsertDevice(device)
	return device
}

// Device holds the name addr and type of device
type Device struct {
	Name     string
	Type     string
	Addr     []byte
	Elements []Element
}

func (d *Device) AddElem(stateType string, db DB) []byte {
	// Check if first elem
	if len(d.Elements) == 0 {
		// Create element with device main addr
		element := Element{
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

func (d *Device) UpdateState(index int, state []byte, db DB) {
	d.Elements[index].State.State = state
	db.UpdateDevice(*d)
}

func (d *Device) UpdateStateUsingAddr(addr, state []byte, db DB) {
	for i, element := range d.Elements {
		if reflect.DeepEqual(element.Addr, addr) {
			d.Elements[i].State.State = state
			db.UpdateDevice(*d)
		}
	}
}

func (d *Device) GetState(index int) State {
	return d.Elements[index].State
}

func (d *Device) GetElemAddr(index int) []byte {
	return d.Elements[index].Addr
}

// Element holds an elements addr and its state
type Element struct {
	Addr  []byte
	State State
}

// State has a value and a type
type State struct {
	State     []byte
	StateType string
}
