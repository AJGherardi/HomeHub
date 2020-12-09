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
	messages chan int
	ctx      context.Context
}

type stateObserver struct {
	groupAddr uint16
	devAddr   uint16
	elemAddr  uint16
	messages  chan string
	ctx       context.Context
}

// Resolver is the root of the schema
type Resolver struct {
	Store              *model.Store
	Controller         mesh.Controller
	Mdns               *zeroconf.Server
	UnprovisionedNodes *[][]byte
	NodeAdded          chan uint16
	StateObservers     []stateObserver
	EventObservers     []eventObserver
	UserPin            int
}

// New returns the servers config
func New(
	store *model.Store,
	controller mesh.Controller,
	nodeAdded chan uint16,
	mdns *zeroconf.Server,
	unprovisionedNodes *[][]byte,
) (generated.Config, func(), func(addr uint16), *Resolver) {
	// Make resolver
	resolver := Resolver{
		Store:              store,
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
			for _, observer := range resolver.StateObservers {
				// Check if observer is closed
				select {
				case <-observer.ctx.Done():
					// If so remove observer and continue
					// utils.Delete(&resolver.StateObservers, i)
					continue
				default:
				}
				device := store.Groups[uint16(observer.groupAddr)].Devices[uint16(observer.devAddr)]
				state := device.GetState(observer.elemAddr)
				observer.messages <- utils.EncodeBase64(state)
			}
		},
		func(addr uint16) {
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
				observer.messages <- int(addr)
			}
		},
		&resolver
}
