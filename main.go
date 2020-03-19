package main

import (
	"context"
	"net/http"

	"github.com/samsarahq/thunder/graphql"
	"github.com/samsarahq/thunder/graphql/introspection"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	netKey            = []byte{0xaf, 0xc3, 0x27, 0x0e, 0xda, 0x88, 0x02, 0xf7, 0x2c, 0x1e, 0x53, 0x24, 0x38, 0xa9, 0x79, 0xeb}
	devicesCollection *mongo.Collection
	keysCollection    *mongo.Collection
	netCollection     *mongo.Collection
	netData           NetData
)

func main() {
	// Get ref to collections
	devicesCollection = getCollection("devices")
	keysCollection = getCollection("keys")
	netCollection = getCollection("net")
	// Delete all objects
	devicesCollection.DeleteMany(context.TODO(), bson.D{})
	keysCollection.DeleteMany(context.TODO(), bson.D{})
	netCollection.DeleteMany(context.TODO(), bson.D{})
	// Add and get net data
	addNetData(netCollection, NetData{NetKey: netKey, NetKeyIndex: []byte{0x00, 0x00}, Flags: []byte{0x00}, IvIndex: []byte{0x00, 0x00, 0x00, 0x00}, NextDevAddr: []byte{0x00, 0x00}})
	netData = getNetData(netCollection)
	// Build schema
	schema := schema()
	introspection.AddIntrospectionToSchema(schema)
	// Serve graphql
	http.Handle("/graphql", graphql.HTTPHandler(schema))
	http.ListenAndServe(":8080", nil)
}
