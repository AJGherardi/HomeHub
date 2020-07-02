package main

import (
	"reflect"

	mesh "github.com/AJGherardi/GoMeshCryptro"
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
