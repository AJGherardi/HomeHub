package main

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	mesh "github.com/AJGherardi/GoMeshController"
	"github.com/AJGherardi/HomeHub/generated"
	"github.com/AJGherardi/HomeHub/graph"
	"github.com/AJGherardi/HomeHub/model"
	"github.com/grandcat/zeroconf"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	mdns               *zeroconf.Server
	unprovisionedNodes [][]byte
	nodeAdded          = make(chan []byte)
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
			unprovisionedNodes = append(unprovisionedNodes, uuid)
		},
		// onNodeAdded
		func(addr []byte) {
			nodeAdded <- addr
		},
	)
	// Check if configured
	if db.GetNetData().ID == primitive.NilObjectID {
		// Setup the mdns service
		mdns, _ = zeroconf.Register("hub", "_alexandergherardi._tcp", "local.", 8080, nil, nil)
	}
	// Serve the schema
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(graph.New(db, controller, nodeAdded, mdns)))
	http.Handle("/", playground.Handler("GraphQL playground", "/graphql"))

	http.Handle("/graphql", srv)
	http.ListenAndServe(":8080", nil)
}
