package graph

import (
	"context"

	mesh "github.com/AJGherardi/GoMeshController"
	"github.com/AJGherardi/HomeHub/generated"
	"github.com/AJGherardi/HomeHub/model"
	"github.com/AJGherardi/HomeHub/utils"
	"github.com/grandcat/zeroconf"
)

type eventObserver struct {
	messages chan string
	ctx      context.Context
}

type stateObserver struct {
	addr     []byte
	messages chan string
	ctx      context.Context
}

// Resolver is the root of the schema
type Resolver struct {
	DB                 model.DB
	Controller         mesh.Controller
	Mdns               *zeroconf.Server
	UnprovisionedNodes *[][]byte
	NodeAdded          chan []byte
	StateObservers     []stateObserver
	EventObservers     []eventObserver
	UserPin            int
}

// New returns the servers config
func New(
	db model.DB,
	controller mesh.Controller,
	nodeAdded chan []byte,
	mdns *zeroconf.Server,
	unprovisionedNodes *[][]byte,
) (generated.Config, func(), func(addr []byte), *Resolver) {
	// Make resolver
	resolver := Resolver{
		DB:                 db,
		Controller:         controller,
		NodeAdded:          nodeAdded,
		Mdns:               mdns,
		UnprovisionedNodes: unprovisionedNodes,
		StateObservers:     make([]stateObserver, 0),
		EventObservers:     make([]eventObserver, 0),
		UserPin:            000000,
	}
	// Make config
	c := generated.Config{
		Resolvers: &resolver,
	}
	// Start updater function
	return c,
		func() {
			// Update observers
			for i, observer := range resolver.StateObservers {
				// Check if observer is closed
				select {
				case <-observer.ctx.Done():
					// If so remove observer and continue
					utils.Delete(&resolver.StateObservers, i)
					continue
				default:
				}
				device := resolver.DB.GetDeviceByElemAddr(observer.addr)
				state := device.GetState(observer.addr)
				observer.messages <- utils.EncodeBase64(state)
			}
		},
		func(addr []byte) {
			// Update observers
			for i, observer := range resolver.EventObservers {
				// Check if observer is closed
				select {
				case <-observer.ctx.Done():
					// If so remove observer and continue
					utils.Delete(&resolver.StateObservers, i)
					continue
				default:
				}
				observer.messages <- utils.EncodeBase64(addr)
			}
		},
		&resolver
}
