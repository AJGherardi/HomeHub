package graph

import (
	"context"

	mesh "github.com/AJGherardi/GoMeshController"
	"github.com/AJGherardi/HomeHub/generated"
	"github.com/AJGherardi/HomeHub/model"
	"github.com/AJGherardi/HomeHub/utils"
	"github.com/grandcat/zeroconf"
)

type observer struct {
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
	Observers          []observer
	UserPin            int
}

// New returns the servers config
func New(
	db model.DB,
	controller mesh.Controller,
	nodeAdded chan []byte,
	mdns *zeroconf.Server,
	unprovisionedNodes *[][]byte,
) (generated.Config, func(), *Resolver) {
	// Make resolver
	resolver := Resolver{
		DB:                 db,
		Controller:         controller,
		NodeAdded:          nodeAdded,
		Mdns:               mdns,
		UnprovisionedNodes: unprovisionedNodes,
		Observers:          make([]observer, 0),
		UserPin:            000000,
	}
	// Make config
	c := generated.Config{
		Resolvers: &resolver,
	}
	// Start updater function
	return c, func() {
		// Update observers
		for i, observer := range resolver.Observers {
			// Check if observer is closed
			select {
			case <-observer.ctx.Done():
				// If so remove observer and continue
				utils.Delete(&resolver.Observers, i)
				continue
			default:
			}
			device := resolver.DB.GetDeviceByElemAddr(observer.addr)
			state := device.GetState(observer.addr)
			observer.messages <- utils.EncodeBase64(state)
		}
	}, &resolver
}
