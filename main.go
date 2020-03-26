package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-ble/ble"
	"github.com/samsarahq/thunder/graphql"
	"github.com/samsarahq/thunder/graphql/introspection"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	devicesCollection *mongo.Collection
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
	devicesCollection = getCollection("devices")
	appKeysCollection = getCollection("appKeys")
	devKeysCollection = getCollection("devKeys")
	netCollection = getCollection("net")
	// Delete all objects
	devicesCollection.DeleteMany(context.TODO(), bson.D{})
	appKeysCollection.DeleteMany(context.TODO(), bson.D{})
	devKeysCollection.DeleteMany(context.TODO(), bson.D{})
	netCollection.DeleteMany(context.TODO(), bson.D{})
	// Add and get net data
	insertNetData(netCollection, NetData{
		ID:          primitive.NewObjectID(),
		NetKey:      []byte{0xaf, 0xc3, 0x27, 0x0e, 0xda, 0x88, 0x02, 0xf7, 0x2c, 0x1e, 0x53, 0x24, 0x38, 0xa9, 0x79, 0xeb},
		NetKeyIndex: []byte{0x00, 0x00},
		Flags:       []byte{0x00},
		IvIndex:     []byte{0x00, 0x00, 0x00, 0x00},
		NextDevAddr: []byte{0x00, 0x01},
		HubSeq:      []byte{0x00, 0x00, 0x00},
	})
	netData = getNetData(netCollection)
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
