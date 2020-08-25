package main

import (
	"reflect"
)

func makeGroup(name string, addr, keyIndex []byte) Group {
	group := Group{
		Name:     name,
		Addr:     addr,
		KeyIndex: keyIndex,
	}
	insertGroup(groupsCollection, group)
	return group
}

// Group holds a collection of devices and its app key
type Group struct {
	Name     string
	KeyIndex []byte
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

func (g *Group) getKeyIndex() []byte {
	return g.KeyIndex
}
