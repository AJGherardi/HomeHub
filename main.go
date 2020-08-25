package main

import (
	"net/http"

	mesh "github.com/AJGherardi/GoMeshController"
	"github.com/grandcat/zeroconf"
	"github.com/samsarahq/thunder/graphql"
	"github.com/samsarahq/thunder/graphql/introspection"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	groupsCollection   *mongo.Collection
	devicesCollection  *mongo.Collection
	netCollection      *mongo.Collection
	mdns               *zeroconf.Server
	unprovisionedNodes [][]byte
	nodeAdded          = make(chan []byte)
	controller         mesh.Controller
)

func main() {
	// Get ref to collections
	groupsCollection = getCollection("groups")
	devicesCollection = getCollection("devices")
	netCollection = getCollection("net")
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
	if getNetData(netCollection).ID == primitive.NilObjectID {
		// Setup the mdns service
		mdns, _ = zeroconf.Register("hub", "_alexandergherardi._tcp", "local.", 8080, nil, nil)
	}
	// Build schema
	schema := schema()
	introspection.AddIntrospectionToSchema(schema)
	// Serve graphql
	http.Handle("/graphql", graphql.HTTPHandler(schema))
	http.ListenAndServe(":8080", nil)
	// connectAndServe(schema)
}
