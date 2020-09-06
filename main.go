package main

import (
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	mesh "github.com/AJGherardi/GoMeshController"
	"github.com/AJGherardi/HomeHub/generated"
	"github.com/AJGherardi/HomeHub/graph"
	"github.com/AJGherardi/HomeHub/model"
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
			stateChanged <- []byte{state}
		},
	)
	// Check if configured
	if db.GetNetData().ID == primitive.NilObjectID {
		// Setup the mdns service
		mdns, _ = zeroconf.Register("hub", "_alexandergherardi._tcp", "local.", 8080, nil, nil)
	}
	// Serve the schema
	srv := handler.New(
		generated.NewExecutableSchema(
			graph.New(db, controller, nodeAdded, mdns, unprovisionedNodes, stateChanged),
		),
	)
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	})
	srv.Use(extension.Introspection{})
	http.Handle("/graphql", srv)
	http.ListenAndServe(":8080", nil)
}
