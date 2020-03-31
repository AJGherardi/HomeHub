package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-ble/ble"
	"github.com/micro/mdns"
	"github.com/samsarahq/thunder/graphql"
	"github.com/samsarahq/thunder/graphql/introspection"
	"go.mongodb.org/mongo-driver/bson"
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
	mdnsServer        *mdns.Server
	write             *ble.Characteristic
	cln               ble.Client
)

func main() {
	// Get ref to collections
	groupsCollection = getCollection("groups")
	devicesCollection = getCollection("devices")
	webKeysCollection = getCollection("webKeys")
	appKeysCollection = getCollection("appKeys")
	devKeysCollection = getCollection("devKeys")
	netCollection = getCollection("net")
	netCollection.DeleteMany(context.TODO(), bson.D{})
	// Check if configured
	if getNetData(netCollection).ID == primitive.NilObjectID {
		// Setup our mdns service
		host, _ := os.Hostname()
		fmt.Println(host)
		mdnsInfo := []string{"Service for the alexander gherardi home hub"}
		mdnsService, _ := mdns.NewMDNSService(host, "_alexandergherardi._tcp", "", "", 8080, nil, mdnsInfo)
		// Create the mDNS server
		mdnsServer, _ = mdns.NewServer(&mdns.Config{Zone: mdnsService})
	} else {
		// Connect and get write characteristic if hub is configured
		cln, write = connectToProxy()
		fmt.Println("con")
	}
	// Build schema
	schema := schema()
	introspection.AddIntrospectionToSchema(schema)
	// Serve graphql
	http.Handle("/graphql", graphql.HTTPHandler(schema, authenticate))
	http.ListenAndServe(":8080", nil)
}
