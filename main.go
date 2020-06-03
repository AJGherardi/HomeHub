package main

import (
	"fmt"
	"net/http"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
	"github.com/grandcat/zeroconf"
	"github.com/samsarahq/thunder/graphql"
	"github.com/samsarahq/thunder/graphql/introspection"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	groupsCollection  *mongo.Collection
	devicesCollection *mongo.Collection
	webKeysCollection *mongo.Collection
	appKeysCollection *mongo.Collection
	devKeysCollection *mongo.Collection
	netCollection     *mongo.Collection
	write             *ble.Characteristic
	read              *ble.Characteristic
	cln               ble.Client
	mdns              *zeroconf.Server
)

func main() {
	// Get ref to collections
	groupsCollection = getCollection("groups")
	devicesCollection = getCollection("devices")
	webKeysCollection = getCollection("webKeys")
	appKeysCollection = getCollection("appKeys")
	devKeysCollection = getCollection("devKeys")
	netCollection = getCollection("net")
	// Get ble device
	d, err := dev.NewDevice("default")
	if err != nil {
		fmt.Println(err)
	}
	ble.SetDefaultDevice(d)
	// Check if configured
	if getNetData(netCollection).ID == primitive.NilObjectID {
		// Setup the mdns service
		mdns, _ = zeroconf.Register("hub", "_alexandergherardi._tcp", "local.", 8080, nil, nil)
	} else {
		// Check if there are no devices
		if len(getDevices(devicesCollection)) != 0 {
			// Connect and get write characteristic if hub is configured
			cln, write, read = connectToProxy()
			go reconnectOnDisconnect(cln.Disconnected())
		}
	}
	// Build schema
	schema := schema()
	introspection.AddIntrospectionToSchema(schema)
	// Serve graphql
	http.Handle("/graphql", graphql.HTTPHandler(schema, authenticate))
	http.ListenAndServe(":8080", nil)
	// connectAndServe(schema)
}
