package main

import (
	"context"

	mesh "github.com/AJGherardi/GoMeshCryptro"
	"go.mongodb.org/mongo-driver/bson"
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
	var result NetData
	cur.Decode(&result)
	return result
}

func insertNetData(collection *mongo.Collection, data NetData) {
	collection.InsertOne(context.TODO(), data)
}

func updateNetData(collection *mongo.Collection, data NetData) {
	collection.UpdateOne(
		context.TODO(),
		bson.M{"_id": data.ID},
		bson.M{"$set": bson.M{
			"netkey":          data.NetKey,
			"netkeyindex":     data.NetKeyIndex,
			"nextappkeyindex": data.NextAppKeyIndex,
			"flags":           data.Flags,
			"ivindex":         data.IvIndex,
			"nextaddr":        data.NextAddr,
			"nextgroupaddr":   data.NextGroupAddr,
			"hubseq":          data.HubSeq,
		}},
	)
}

func getGroups(collection *mongo.Collection) []Group {
	var groups []Group
	// Get all Devices
	cur, _ := collection.Find(context.TODO(), bson.D{})
	// Deserialize into array of Groups
	for cur.Next(context.TODO()) {
		var result Group
		cur.Decode(&result)
		// Add to array
		groups = append(groups, result)
	}
	return groups
}

func getGroupByAddr(collection *mongo.Collection, addr []byte) Group {
	var group Group
	result := collection.FindOne(context.TODO(), bson.M{"addr": addr})
	result.Decode(&group)
	return group
}

func insertGroup(collection *mongo.Collection, group Group) {
	collection.InsertOne(context.TODO(), group)
}

func updateGroup(collection *mongo.Collection, group Group) {
	collection.UpdateOne(
		context.TODO(),
		bson.M{"addr": group.Addr},
		bson.M{"$set": bson.M{
			"aid":      group.Aid,
			"name":     group.Name,
			"devaddrs": group.DevAddrs,
		}},
	)
}

func getDevices(collection *mongo.Collection) []Device {
	var devices []Device
	// Get all Devices
	cur, _ := collection.Find(context.TODO(), bson.D{})
	// Deserialize into array of devices
	for cur.Next(context.TODO()) {
		var result Device
		cur.Decode(&result)
		// Add to array
		devices = append(devices, result)
	}
	return devices
}

func insertDevice(collection *mongo.Collection, device Device) {
	collection.InsertOne(context.TODO(), device)
}

func updateDevice(collection *mongo.Collection, data Device) {
	collection.UpdateOne(
		context.TODO(),
		bson.M{"addr": data.Addr},
		bson.M{"$set": bson.M{
			"name":     data.Name,
			"addr":     data.Addr,
			"type":     data.Type,
			"seq":      data.Seq,
			"elements": data.Elements,
		}},
	)
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

func getAppKeyByAid(collection *mongo.Collection, aid []byte) mesh.AppKey {
	var key mesh.AppKey
	result := collection.FindOne(context.TODO(), bson.M{"aid": aid})
	result.Decode(&key)
	return key
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
		var result mesh.DevKey
		cur.Decode(&result)
		// Add to array
		keys = append(keys, result)
	}
	return keys
}

func insertDevKey(collection *mongo.Collection, key mesh.DevKey) {
	collection.InsertOne(context.TODO(), key)
}
