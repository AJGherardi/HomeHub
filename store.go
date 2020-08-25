package main

import (
	"context"
	"reflect"

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
			"nextappkeyindex": data.NextAppKeyIndex,
			"nextgroupaddr":   data.NextGroupAddr,
			"webkeys":         data.WebKeys,
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

func getGroupByDevAddr(collection *mongo.Collection, addr []byte) Group {
	// Get all Devices
	cur, _ := collection.Find(context.TODO(), bson.D{})
	// Deserialize into array of Groups
	for cur.Next(context.TODO()) {
		var result Group
		cur.Decode(&result)
		for _, devAddr := range result.DevAddrs {
			if reflect.DeepEqual(devAddr, addr) {
				return result
			}
		}
	}
	return Group{}
}

func insertGroup(collection *mongo.Collection, group Group) {
	collection.InsertOne(context.TODO(), group)
}

func updateGroup(collection *mongo.Collection, group Group) {
	collection.UpdateOne(
		context.TODO(),
		bson.M{"addr": group.Addr},
		bson.M{"$set": bson.M{
			"keyIndex": group.KeyIndex,
			"name":     group.Name,
			"devaddrs": group.DevAddrs,
		}},
	)
}

func deleteGroup(collection *mongo.Collection, addr []byte) {
	collection.DeleteOne(
		context.TODO(),
		bson.M{"addr": addr},
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

func getDeviceByElemAddr(collection *mongo.Collection, elemAddr []byte) Device {
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
	for _, device := range devices {
		for _, element := range device.Elements {
			if reflect.DeepEqual(element.Addr, elemAddr) {
				return device
			}
		}
	}
	return Device{}
}

func getDeviceByAddr(collection *mongo.Collection, addr []byte) Device {
	var device Device
	result := collection.FindOne(context.TODO(), bson.M{"addr": addr})
	result.Decode(&device)
	return device
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
			"elements": data.Elements,
		}},
	)
}

func deleteDevice(collection *mongo.Collection, addr []byte) {
	collection.DeleteOne(
		context.TODO(),
		bson.M{"addr": addr},
	)
}
