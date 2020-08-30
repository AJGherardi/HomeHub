package model

import (
	"reflect"

	"github.com/AJGherardi/HomeHub/utils"
)

// MakeGroup makes a new group with the given addr
func MakeGroup(name string, addr, keyIndex []byte, db DB) Group {
	group := Group{
		Name:     name,
		Addr:     addr,
		KeyIndex: keyIndex,
	}
	db.InsertGroup(group)
	return group
}

// Group holds a collection of devices and its app key
type Group struct {
	Name     string
	KeyIndex []byte
	Addr     []byte
	DevAddrs [][]byte
}

// AddDevice adds a device to the group
func (g *Group) AddDevice(addr []byte, db DB) {
	g.DevAddrs = append(g.DevAddrs, addr)
	db.UpdateGroup(*g)
}

// RemoveDevice removes the device from the group
func (g *Group) RemoveDevice(devAddr []byte, db DB) {
	for i, addr := range g.DevAddrs {
		if reflect.DeepEqual(addr, devAddr) {
			g.DevAddrs = utils.RemoveDevAddr(g.DevAddrs, i)
		}
	}
	db.UpdateGroup(*g)
}

// GetDevAddrs returns all the devices addrs in the group
func (g *Group) GetDevAddrs() [][]byte {
	return g.DevAddrs
}

// GetKeyIndex returns the app key index
func (g *Group) GetKeyIndex() []byte {
	return g.KeyIndex
}
