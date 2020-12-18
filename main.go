package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	mesh "github.com/AJGherardi/GoMeshController"
	"github.com/AJGherardi/HomeHub/cmd"
	"github.com/AJGherardi/HomeHub/generated"
	"github.com/AJGherardi/HomeHub/graph"
	"github.com/AJGherardi/HomeHub/model"
	"github.com/AJGherardi/HomeHub/utils"
	"github.com/gorilla/websocket"
	"github.com/grandcat/zeroconf"
)

var (
	mdns               *zeroconf.Server
	unprovisionedNodes = new([][]byte)
	nodeAdded          = make(chan uint16)
	stateChanged       = make(chan []byte)
	controller         mesh.Controller
	store              model.Store
)

func main() {
	store := model.Store{
		Groups: map[uint16]*model.Group{},
	}
	// Open Mesh Controller and defer close
	controller, err := mesh.Open()
	if err != nil {
		panic("Can not open mesh controller")
	}
	defer controller.Close()
	// Generate a cert
	if _, err := os.Stat("home.data"); err != nil {
		if os.IsNotExist(err) {
			utils.WriteCert()
		}
	}
	// Check if configured
	if !utils.CheckIfConfigured() {
		mdns, err = zeroconf.Register("unprovisioned", "_alexandergherardi._tcp", "local.", 8080, nil, nil)
		if err != nil {
			panic("Can not register mdns service")
		}
	} else {
		mdns, err = zeroconf.Register("hub", "_alexandergherardi._tcp", "local.", 8080, nil, nil)
		if err != nil {
			panic("Can not register mdns service")
		}
		store, _ = cmd.ReadFromFile()
		go cmd.SaveStore(&store)
	}
	// Serve the schema
	schema, updateState, publishEvents, resolver := graph.New(&store, controller, nodeAdded, mdns, unprovisionedNodes)
	srv := handler.New(
		generated.NewExecutableSchema(
			schema,
		),
	)
	// Register read functions
	go func() {
		err := controller.Read(
			// onSetupStatus
			func() {},
			// onAddKeyStatus
			func(appIdx uint16) {},
			// onUnprovisionedBeacon
			func(uuid []byte) {
				*unprovisionedNodes = append(*unprovisionedNodes, uuid)
			},
			// onNodeAdded
			func(addr uint16) {
				nodeAdded <- addr
			},
			// onState
			func(addr uint16, state byte) {
				cmd.UpdateState(&store, addr, state)
				// Push new state
				updateState()
			},
			// onEvent
			func(addr uint16) {
				publishEvents(addr)
			},
		)
		if err != nil {
			panic("Can not read from mesh controller")
		}
	}()

	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		InitFunc: func(ctx context.Context, initPayload transport.InitPayload) (context.Context, error) {
			// If not configured auth is not required
			if !utils.CheckIfConfigured() {
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
			verify := store.NetData.CheckWebKey(webKey)
			// If valid continue
			if verify {
				return ctx, nil
			}
			// Else error
			return ctx, errors.New("Bad webKey")
		},
	})
	srv.AddTransport(transport.POST{})
	srv.Use(extension.Introspection{})
	http.Handle("/graphql", srv)
	err = http.ListenAndServeTLS("", "cert.pem", "key.pem", nil)
	log.Fatal(err)
}
