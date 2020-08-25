package main

import (
	"reflect"
)

func makeDevice(name, deviceType string, addr []byte) Device {
	device := Device{
		Name: name,
		Type: deviceType,
		Addr: addr,
	}
	insertDevice(devicesCollection, device)
	return device
}

// Device holds the name addr and type of device
type Device struct {
	Name     string
	Type     string
	Addr     []byte
	Elements []Element
}

func (d *Device) addElem(stateType string) []byte {
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
	addr := incrementAddr(d.Elements[len(d.Elements)-1].Addr)
	element := Element{
		Addr: addr,
		State: State{
			State:     []byte{0x00},
			StateType: stateType,
		},
	}
	d.Elements = append(d.Elements, element)
	updateDevice(devicesCollection, *d)
	return addr
}

func (d *Device) updateState(index int, state []byte) {
	d.Elements[index].State.State = state
	updateDevice(devicesCollection, *d)
}

func (d *Device) updateStateUsingAddr(addr, state []byte) {
	for i, element := range d.Elements {
		if reflect.DeepEqual(element.Addr, addr) {
			d.Elements[i].State.State = state
			updateDevice(devicesCollection, *d)
		}
	}
}

func (d *Device) getState(index int) State {
	return d.Elements[index].State
}

func (d *Device) getElemAddr(index int) []byte {
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
