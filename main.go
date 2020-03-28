package main

import (
	"fmt"
	"net/http"

	"github.com/go-ble/ble"
	"github.com/samsarahq/thunder/graphql"
	"github.com/samsarahq/thunder/graphql/introspection"
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
	cln               ble.Client
	netData           NetData
	messages          = make(chan []byte)
)

func main() {
	// Get ref to collections
	groupsCollection = getCollection("groups")
	devicesCollection = getCollection("devices")
	webKeysCollection = getCollection("webKeys")
	appKeysCollection = getCollection("appKeys")
	devKeysCollection = getCollection("devKeys")
	netCollection = getCollection("net")
	// Connect and get write characteristic
	cln, write = connectToProxy()
	fmt.Println("con")
	// Build schema
	schema := schema()
	introspection.AddIntrospectionToSchema(schema)
	// Serve graphql
	http.Handle("/graphql", graphql.HTTPHandler(schema))
	http.ListenAndServe(":8080", nil)
}
