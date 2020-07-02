package main

import (
	"reflect"

	mesh "github.com/AJGherardi/GoMeshCryptro"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func makeGroup(name string, addr []byte, appKey mesh.AppKey) Group {
	group := Group{
		Name:   name,
		Addr:   addr,
		AppKey: appKey,
	}
	insertGroup(groupsCollection, group)
	return group
}

// Group holds a collection of devices and its app key id
type Group struct {
	Name     string
	AppKey   mesh.AppKey
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

func (g *Group) getAppKey() mesh.AppKey {
	return g.AppKey
}

func makeDevice(name, deviceType string, addr []byte, devKey mesh.DevKey) Device {
	device := Device{
		Name:   name,
		Type:   deviceType,
		Addr:   addr,
		DevKey: devKey,
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
	DevKey   mesh.DevKey
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
func (d *Device) getDevKey(index int) mesh.DevKey {
	return d.DevKey
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
