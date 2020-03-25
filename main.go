package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-ble/ble"
	"github.com/samsarahq/thunder/graphql"
	"github.com/samsarahq/thunder/graphql/introspection"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	netKey            = []byte{0xaf, 0xc3, 0x27, 0x0e, 0xda, 0x88, 0x02, 0xf7, 0x2c, 0x1e, 0x53, 0x24, 0x38, 0xa9, 0x79, 0xeb}
	devicesCollection *mongo.Collection
	appKeysCollection *mongo.Collection
	devKeysCollection *mongo.Collection
	netCollection     *mongo.Collection
	write             *ble.Characteristic
	cln               ble.Client
	netData           NetData
	messages          = make(chan []byte)
	// devKey            = []byte{0x96, 0x4a, 0xf6, 0xfc, 0x03, 0x38, 0x8c, 0x73, 0xea, 0xff, 0x94, 0x61, 0x57, 0xff, 0x66, 0x01}
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
	insertNetData(netCollection, NetData{NetKey: netKey, NetKeyIndex: []byte{0x00, 0x00}, Flags: []byte{0x00}, IvIndex: []byte{0x00, 0x00, 0x00, 0x00}, NextDevAddr: []byte{0x00, 0x01}})
	netData = getNetData(netCollection)
	// Add and get App Keys
	// insertAppKey(appKeysCollection, mesh.AppKey{Aid: []byte{0x21}, Key: netKey, KeyIndex: []byte{0x01, 0x02}})
	// getAppKeys(appKeysCollection)
	// Add and get Dev Keys
	// insertDevKey(devKeysCollection, mesh.DevKey{Addr: []byte{0x00, 0x01}, Key: netKey})
	getDevKeys(devKeysCollection)
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
