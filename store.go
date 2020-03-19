package main

import (
	"context"

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

func addNetData(collection *mongo.Collection, data NetData) {
	collection.InsertOne(context.TODO(), data)
}

func getDevices(collection *mongo.Collection) []Device {
	var devices []Device
	// Get all keys
	cur, _ := collection.Find(context.TODO(), bson.D{})
	// Deserialize into array of devices
	for cur.Next(context.TODO()) {
		var result bson.M
		cur.Decode(&result)
		name := result["name"].(string)
		addr := result["addr"].(string)
		devType := result["type"].(string)
		// Add to array
		devices = append(devices, Device{Addr: addr, Name: name, Type: devType})
	}
	return devices
}

func addDevice(collection *mongo.Collection, device Device) {
	collection.InsertOne(context.TODO(), device)
}

func getAppKeys(collection *mongo.Collection) []AppKey {
	var keys []AppKey
	// Get all keys
	cur, _ := collection.Find(context.TODO(), bson.D{})
	// Deserialize into array of app keys
	for cur.Next(context.TODO()) {
		var result bson.M
		cur.Decode(&result)
		ID := result["id"].(primitive.Binary).Data
		key := result["key"].(primitive.Binary).Data
		keyType := result["type"].(bool)
		// Add to arrays
		keys = append(keys, AppKey{ID: ID, Key: key, Type: KeyType(keyType)})
	}
	return keys
}

func addAppKey(collection *mongo.Collection, key AppKey) {
	collection.InsertOne(context.TODO(), key)
}
