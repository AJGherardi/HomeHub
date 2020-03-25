package main

import (
	"context"

	mesh "github.com/AJGherardi/GoMeshCryptro"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func getCollection(database string) *mongo.Collection {
	// Get client
	client, _ := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	// Connect with timeout
	client.Connect(context.TODO())
	// Test using ping
	client.Ping(context.TODO(), readpref.Primary())
	// Get collection ref
	collection := client.Database("main").Collection(database)
	return collection
}

func getNetData(collection *mongo.Collection) NetData {
	cur, _ := collection.Find(context.TODO(), bson.D{})
	// Deserialize first result
	cur.Next(context.TODO())
	var result bson.M
	cur.Decode(&result)
	netKey := result["netkey"].(primitive.Binary).Data
	netKeyIndex := result["netkeyindex"].(primitive.Binary).Data
	flags := result["flags"].(primitive.Binary).Data
	ivIndex := result["ivindex"].(primitive.Binary).Data
	nextDevAddr := result["nextdevaddr"].(primitive.Binary).Data
	// Build net data struct
	netData := NetData{NetKey: netKey, NetKeyIndex: netKeyIndex, Flags: flags, IvIndex: ivIndex, NextDevAddr: nextDevAddr}
	return netData
}

func insertNetData(collection *mongo.Collection, data NetData) {
	collection.InsertOne(context.TODO(), data)
}

func getDevices(collection *mongo.Collection) []Device {
	var devices []Device
	// Get all Devices
	cur, _ := collection.Find(context.TODO(), bson.D{})
	// Deserialize into array of devices
	for cur.Next(context.TODO()) {
		var result bson.M
		cur.Decode(&result)
		name := result["name"].(string)
		devType := result["type"].(string)
		addr := result["addr"].(primitive.Binary).Data
		// Add to array
		devices = append(devices, Device{Addr: addr, Name: name, Type: devType})
	}
	return devices
}

func insertDevice(collection *mongo.Collection, device Device) {
	collection.InsertOne(context.TODO(), device)
}

func getAppKeys(collection *mongo.Collection) []mesh.AppKey {
	var keys []mesh.AppKey
	// Get all keys
	cur, _ := collection.Find(context.TODO(), bson.D{})
	// Deserialize into array of app keys
	for cur.Next(context.TODO()) {
		var result mesh.AppKey
		cur.Decode(&result)

		// Add to array
		keys = append(keys, result)
	}
	return keys
}

func insertAppKey(collection *mongo.Collection, key mesh.AppKey) {
	collection.InsertOne(context.TODO(), key)
}

func getDevKeys(collection *mongo.Collection) []mesh.DevKey {
	var keys []mesh.DevKey
	// Get all keys
	cur, _ := collection.Find(context.TODO(), bson.D{})
	// Deserialize into array of dev keys
	for cur.Next(context.TODO()) {
		var result bson.M
		cur.Decode(&result)
		addr := result["addr"].(primitive.Binary).Data
		key := result["key"].(primitive.Binary).Data
		// Add to array
		keys = append(keys, mesh.DevKey{Addr: addr, Key: key})
	}
	return keys
}

func insertDevKey(collection *mongo.Collection, key mesh.DevKey) {
	collection.InsertOne(context.TODO(), key)
}
