package graph

import (
	"context"
	"reflect"

	mesh "github.com/AJGherardi/GoMeshController"
	"github.com/AJGherardi/HomeHub/generated"
	"github.com/AJGherardi/HomeHub/model"
	"github.com/AJGherardi/HomeHub/utils"
	"github.com/grandcat/zeroconf"
)

type observer struct {
	addr     []byte
	messages chan []*model.Device
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
}

// New returns the servers config
func New(
	db model.DB,
	controller mesh.Controller,
	nodeAdded chan []byte,
	mdns *zeroconf.Server,
	unprovisionedNodes *[][]byte,
	stateChanged chan []byte,
) generated.Config {
	// Make resolver
	resolver := Resolver{
		DB:                 db,
		Controller:         controller,
		NodeAdded:          nodeAdded,
		Mdns:               mdns,
		UnprovisionedNodes: unprovisionedNodes,
		Observers:          make([]observer, 0),
	}
	// Make config
	c := generated.Config{
		Resolvers: &resolver,
	}
	// Start updater function
	go func() {
		for {
			// Wait for state changed
			<-stateChanged
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
				group := resolver.DB.GetGroupByAddr(observer.addr)
				devicePointers := make([]*model.Device, 0)
				for _, device := range resolver.DB.GetDevices() {
					for _, devAddr := range group.DevAddrs {
						if reflect.DeepEqual(devAddr, device.Addr) {
							devicePointers = append(devicePointers, &device)
						}
					}
				}
				observer.messages <- devicePointers
			}
		}
	}()
	return c
}
