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
	Scenes   []Scene
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

// GetScenes returns the all the scenes in a group
func (g *Group) GetScenes() []Scene {
	return g.Scenes
}

// AddScene adds a scene to a the group
func (g *Group) AddScene(name string, number []byte, db DB) {
	g.Scenes = append(g.Scenes, Scene{Name: name, Number: number})
	db.UpdateGroup(*g)
}

// DeleteScene removes a scene from the group
func (g *Group) DeleteScene(number []byte, db DB) {
	for i, scene := range g.Scenes {
		if reflect.DeepEqual(number, scene.Number) {
			utils.Delete(&g.Scenes, i)
		}
	}
	db.UpdateGroup(*g)
}

// Scene holds the scenes name and number
type Scene struct {
	Name   string
	Number []byte
}
