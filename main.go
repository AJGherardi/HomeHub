package main

import (
	"context"
	"net/http"

	mesh "github.com/AJGherardi/GoMeshCryptro"
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
	netData           NetData
)

func main() {
	// Get ref to collections
	devicesCollection = getCollection("devices")
	appKeysCollection := getCollection("appKeys")
	devKeysCollection = getCollection("devKeys")
	netCollection = getCollection("net")
	// Delete all objects
	devicesCollection.DeleteMany(context.TODO(), bson.D{})
	appKeysCollection.DeleteMany(context.TODO(), bson.D{})
	devKeysCollection.DeleteMany(context.TODO(), bson.D{})
	netCollection.DeleteMany(context.TODO(), bson.D{})
	// Add and get net data
	addNetData(netCollection, NetData{NetKey: netKey, NetKeyIndex: []byte{0x00, 0x00}, Flags: []byte{0x00}, IvIndex: []byte{0x00, 0x00, 0x00, 0x00}, NextDevAddr: []byte{0x00, 0x00}})
	netData = getNetData(netCollection)
	// Add and get App Keys
	addAppKey(appKeysCollection, mesh.AppKey{Aid: []byte{0x21}, Key: netKey, KeyIndex: []byte{0x01, 0x02}})
	getAppKeys(appKeysCollection)
	// Add and get Dev Keys
	addDevKey(devKeysCollection, mesh.DevKey{Addr: []byte{0x00, 0x01}, Key: netKey})
	getDevKeys(devKeysCollection)
	// Build schema
	schema := schema()
	introspection.AddIntrospectionToSchema(schema)
	// Serve graphql
	http.Handle("/graphql", graphql.HTTPHandler(schema))
	http.ListenAndServe(":8080", nil)
}
