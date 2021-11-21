package graph

import (
	"context"

	"github.com/AJGherardi/HomeHub/cmd"
	"github.com/AJGherardi/HomeHub/generated"
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
	Network        *cmd.Network
	Mdns           *zeroconf.Server
	StateObservers []stateObserver
	EventObservers []eventObserver
	UserPin        int
}

// New returns the servers config
func New(
	network *cmd.Network,
	mdns *zeroconf.Server,
) (generated.Config, *Resolver, func(), func(addr uint16)) {
	// Make resolver
	resolver := Resolver{
		Network:        network,
		Mdns:           mdns,
		StateObservers: make([]stateObserver, 0),
		EventObservers: make([]eventObserver, 0),
		UserPin:        000000,
	}
	// Make config
	c := generated.Config{
		Resolvers: &resolver,
	}
	// Start updater function
	return c, &resolver,
		func() {
			// Update observers
			for _, observer := range resolver.StateObservers {
				// Check if observer is closed
				select {
				case <-observer.ctx.Done():
					// If so remove observer and continue
					utils.Delete(&resolver.StateObservers, i)
					continue
				default:
				}
				state, _ := network.ReadState(observer.groupAddr, observer.devAddr, observer.elemAddr)
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
		}
}
