package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/AJGherardi/HomeHub/cmd"
	"github.com/AJGherardi/HomeHub/generated"
	"github.com/AJGherardi/HomeHub/graph"
	"github.com/AJGherardi/HomeHub/utils"
	"github.com/gorilla/websocket"
	"github.com/grandcat/zeroconf"
)

var (
	mdns *zeroconf.Server
)

func main() {
	// Make a network and defer close
	network, read, write := cmd.MakeNetwork()
	defer network.Close()
	// Check if configured
	if !network.CheckIfConfigured() {
		// Generate a cert
		utils.WriteCert()
		mdns = registerMDNS("unprovisioned")
	} else {
		mdns = registerMDNS("hub")
	}
	// Create the schema
	schema, resolver, updateState, publishEvents := graph.New(&network, mdns)
	// Register read function
	go read(updateState, publishEvents)
	// Start save store goroutine
	go write()
	// Serve the schema
	serveAPI(schema, resolver)
}

func serveAPI(schema generated.Config, resolver *graph.Resolver) {
	// Create server using schema
	srv := handler.New(
		generated.NewExecutableSchema(
			schema,
		),
	)
	srv.Use(extension.Introspection{})
	// Add websocket transport with auth layer
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		InitFunc: func(ctx context.Context, initPayload transport.InitPayload) (context.Context, error) {
			// If not configured auth is not required
			if !resolver.Network.CheckIfConfigured() {
				return ctx, nil
			}
			// If pin is available check it
			if resolver.UserPin != 000000 {
				if initPayload["pin"] != nil {
					if initPayload["pin"].(float64) == float64(resolver.UserPin) {
						return ctx, nil
					}
				}
			}
			// Check web key
			webKey := utils.DecodeBase64(initPayload["webKey"].(string))
			verify := resolver.Network.Store.CheckWebKey(webKey)
			// If valid continue
			if verify {
				return ctx, nil
			}
			// Else error
			return ctx, errors.New("Bad webKey")
		},
	})
	// Serve handler
	http.Handle("/graphql", srv)
	err := http.ListenAndServeTLS("", "cert.pem", "key.pem", nil)
	log.Fatal(err)
}

func registerMDNS(name string) *zeroconf.Server {
	mdns, err := zeroconf.Register(name, "_alexandergherardi._tcp", "local.", 8080, nil, nil)
	if err != nil {
		panic("Can not register mdns service")
	}
	return mdns
}
