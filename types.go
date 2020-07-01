package main

import (
	"reflect"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func makeGroup(name string, aid, addr []byte) Group {
	group := Group{
		Name: name,
		Aid:  aid,
		Addr: addr,
	}
	insertGroup(groupsCollection, group)
	return group
}

// Group holds a collection of devices and its app key id
type Group struct {
	Name     string
	Aid      []byte
	Addr     []byte
	DevAddrs [][]byte
}

func (g *Group) addDevice(addr []byte) {
	g.DevAddrs = append(g.DevAddrs, addr)
	updateGroup(groupsCollection, *g)
}

func (g *Group) removeDevice(devAddr []byte) {
	for i, addr := range g.DevAddrs {
		if reflect.DeepEqual(addr, devAddr) {
			g.DevAddrs = removeDevAddr(g.DevAddrs, i)
		}
	}
	updateGroup(groupsCollection, *g)
}

func (g *Group) getDevAddrs() [][]byte {
	return g.DevAddrs
}

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
	Seq      []byte
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

func (d *Device) updateSeq(seq []byte) {
	d.Seq = seq
	updateDevice(devicesCollection, *d)
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

// NetData used for sending msgs and adding new devices
type NetData struct {
	ID              primitive.ObjectID `bson:"_id"`
	NetKey          []byte
	NetKeyIndex     []byte
	NextAppKeyIndex []byte
	Flags           []byte
	IvIndex         []byte
	NextAddr        []byte
	NextGroupAddr   []byte
	HubSeq          []byte
	WebKeys         [][]byte
}
