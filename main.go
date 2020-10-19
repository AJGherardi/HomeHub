package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	mesh "github.com/AJGherardi/GoMeshController"
	"github.com/AJGherardi/HomeHub/generated"
	"github.com/AJGherardi/HomeHub/graph"
	"github.com/AJGherardi/HomeHub/model"
	"github.com/AJGherardi/HomeHub/utils"
	"github.com/gorilla/websocket"
	"github.com/grandcat/zeroconf"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	mdns               *zeroconf.Server
	unprovisionedNodes = new([][]byte)
	nodeAdded          = make(chan []byte)
	stateChanged       = make(chan []byte)
	controller         mesh.Controller
)

func main() {
	// Get db
	db := model.OpenDB()
	// Open Mesh Controller and defer close
	controller = mesh.Open()
	defer controller.Close()
	// Check if configured
	if db.GetNetData().ID == primitive.NilObjectID {
		// Setup the mdns service
		mdns, _ = zeroconf.Register("unprovisioned", "_alexandergherardi._tcp", "local.", 8080, nil, nil)
	} else {
		mdns, _ = zeroconf.Register("hub", "_alexandergherardi._tcp", "local.", 8080, nil, nil)
	}
	// Serve the schema
	schema, updateState, publishEvents, resolver := graph.New(db, controller, nodeAdded, mdns, unprovisionedNodes)
	srv := handler.New(
		generated.NewExecutableSchema(
			schema,
		),
	)
	// Register read functions
	go controller.Read(
		// onSetupStatus
		func() {},
		// onAddKeyStatus
		func(appIdx []byte) {},
		// onUnprovisionedBeacon
		func(uuid []byte) {
			*unprovisionedNodes = append(*unprovisionedNodes, uuid)
		},
		// onNodeAdded
		func(addr []byte) {
			nodeAdded <- addr
		},
		// onState
		func(addr []byte, state byte) {
			// Update device state
			device := db.GetDeviceByElemAddr(addr)
			device.UpdateState(addr, []byte{state}, db)
			updateState()
		},
		// onEvent
		func(addr []byte) {
			publishEvents(addr)
		},
	)
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		InitFunc: func(ctx context.Context, initPayload transport.InitPayload) (context.Context, error) {
			netData := db.GetNetData()
			// If not configured auth is not required
			if netData.ID == primitive.NilObjectID {
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
			verify := netData.CheckWebKey(webKey)
			// If valid continue
			if verify {
				return ctx, nil
			}
			// Else error
			return ctx, errors.New("Bad webKey")
		},
	})
	srv.Use(extension.Introspection{})
	http.Handle("/graphql", srv)
	http.ListenAndServe(":8080", nil)
}
