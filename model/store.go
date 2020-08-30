package model

import (
	"context"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type DB struct {
	GroupsCollection  *mongo.Collection
	DevicesCollection *mongo.Collection
	NetCollection     *mongo.Collection
}

func OpenDB() DB {
	groupsCollection := getCollection("groups")
	devicesCollection := getCollection("devices")
	netCollection := getCollection("net")
	db := DB{
		GroupsCollection:  groupsCollection,
		DevicesCollection: devicesCollection,
		NetCollection:     netCollection,
	}
	return db
}

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

func (db *DB) GetNetData() NetData {
	cur, _ := db.NetCollection.Find(context.TODO(), bson.D{})
	// Deserialize first result
	cur.Next(context.TODO())
	var result NetData
	cur.Decode(&result)
	return result
}

func (db *DB) InsertNetData(data NetData) {
	db.NetCollection.InsertOne(context.TODO(), data)
}

func (db *DB) UpdateNetData(data NetData) {
	db.NetCollection.UpdateOne(
		context.TODO(),
		bson.M{"_id": data.ID},
		bson.M{"$set": bson.M{
			"nextappkeyindex": data.NextAppKeyIndex,
			"nextgroupaddr":   data.NextGroupAddr,
			"webkeys":         data.WebKeys,
		}},
	)
}

func (db *DB) GetGroups() []Group {
	var groups []Group
	// Get all Devices
	cur, _ := db.GroupsCollection.Find(context.TODO(), bson.D{})
	// Deserialize into array of Groups
	for cur.Next(context.TODO()) {
		var result Group
		cur.Decode(&result)
		// Add to array
		groups = append(groups, result)
	}
	return groups
}

func (db *DB) GetGroupByAddr(addr []byte) Group {
	var group Group
	result := db.GroupsCollection.FindOne(context.TODO(), bson.M{"addr": addr})
	result.Decode(&group)
	return group
}

func (db *DB) GetGroupByDevAddr(addr []byte) Group {
	// Get all Devices
	cur, _ := db.GroupsCollection.Find(context.TODO(), bson.D{})
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

func (db *DB) InsertGroup(group Group) {
	db.GroupsCollection.InsertOne(context.TODO(), group)
}

func (db *DB) UpdateGroup(group Group) {
	db.GroupsCollection.UpdateOne(
		context.TODO(),
		bson.M{"addr": group.Addr},
		bson.M{"$set": bson.M{
			"keyIndex": group.KeyIndex,
			"name":     group.Name,
			"devaddrs": group.DevAddrs,
		}},
	)
}

func (db *DB) DeleteGroup(addr []byte) {
	db.GroupsCollection.DeleteOne(
		context.TODO(),
		bson.M{"addr": addr},
	)
}

func (db *DB) GetDevices() []Device {
	var devices []Device
	// Get all Devices
	cur, _ := db.DevicesCollection.Find(context.TODO(), bson.D{})
	// Deserialize into array of devices
	for cur.Next(context.TODO()) {
		var result Device
		cur.Decode(&result)
		// Add to array
		devices = append(devices, result)
	}
	return devices
}

func (db *DB) GetDeviceByElemAddr(elemAddr []byte) Device {
	var devices []Device
	// Get all Devices
	cur, _ := db.DevicesCollection.Find(context.TODO(), bson.D{})
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

func (db *DB) GetDeviceByAddr(addr []byte) Device {
	var device Device
	result := db.DevicesCollection.FindOne(context.TODO(), bson.M{"addr": addr})
	result.Decode(&device)
	return device
}

func (db *DB) InsertDevice(device Device) {
	db.DevicesCollection.InsertOne(context.TODO(), device)
}

func (db *DB) UpdateDevice(data Device) {
	db.DevicesCollection.UpdateOne(
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

func (db *DB) DeleteDevice(addr []byte) {
	db.DevicesCollection.DeleteOne(
		context.TODO(),
		bson.M{"addr": addr},
	)
}

func (db *DB) DeleteAll() {
	db.GroupsCollection.DeleteMany(context.TODO(), bson.D{})
	db.DevicesCollection.DeleteMany(context.TODO(), bson.D{})
	db.NetCollection.DeleteMany(context.TODO(), bson.D{})
}
