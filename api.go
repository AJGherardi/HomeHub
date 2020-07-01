package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	mesh "github.com/AJGherardi/GoMeshCryptro"
	"github.com/grandcat/zeroconf"
	"github.com/samsarahq/thunder/graphql"
	"github.com/samsarahq/thunder/graphql/schemabuilder"
	"github.com/samsarahq/thunder/reactive"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func registerQuery(schema *schemabuilder.Schema) {
	obj := schema.Query()
	obj.FieldFunc("listDevices", func() ([]Device, error) {
		return getDevices(devicesCollection), nil
	})
	obj.FieldFunc("listGroups", func() ([]Group, error) {
		return getGroups(groupsCollection), nil
	})
	obj.FieldFunc("availableDevices", func() ([]string, error) {
		nodes := findDevices()
		return nodes, nil
	})
	obj.FieldFunc("getState", func(args struct {
		DevAddr    string
		ElemNumber int64
	}) (State, error) {
		devAddr := decodeBase64(args.DevAddr)
		device := getDeviceByAddr(devicesCollection, devAddr)
		return device.getState(int(args.ElemNumber)), nil
	})
}

func registerMutation(schema *schemabuilder.Schema) {
	obj := schema.Mutation()
	obj.FieldFunc("addDevice", func(args struct {
		Name    string
		Addr    string
		DevAddr string
	}) (Device, error) {
		// Get net data
		netData := getNetData(netCollection)
		// Connect to unprovisioned device
		cln, write, read = connectToUnprovisioned(args.DevAddr)
		// Provision device
		devKey := provisionDevice(
			cln,
			write,
			netData.NetKey,
			netData.NetKeyIndex,
			netData.Flags,
			netData.IvIndex,
			netData.NextAddr,
		)
		cln.CancelConnection()
		cln, write, read = nil, nil, nil
		time.Sleep(1 * time.Second)
		// Connect to proxy node
		cln, write, read = connectToProxy()
		go reconnectOnDisconnect(cln.Disconnected())
		// Create device object
		device := makeDevice(args.Name, "2PowerSwitch", netData.NextAddr)
		// Get group
		groupAddr := decodeBase64(args.Addr)
		group := getGroupByAddr(groupsCollection, groupAddr)
		// Get app key
		appKey := getAppKeyByAid(appKeysCollection, group.Aid)
		// Insert the dev key
		insertDevKey(devKeysCollection, mesh.DevKey{Addr: device.Addr, Key: devKey})
		// Send app key add
		addPayload := appKeyAdd(netData.NetKeyIndex, appKey.KeyIndex, appKey.Key)
		sendMsg(device.Addr, devKey, addPayload, mesh.DevMsg)
		// Get model id
		if true {
			// Set type and add elements
			elemAddr1 := device.addElem("onoff")
			elemAddr2 := device.addElem("onoff")
			// Send app key bind for onoff
			bindPayload1 := appKeyBind(elemAddr1, appKey.KeyIndex, []byte{0x10, 0x00})
			sendMsg(device.Addr, devKey, bindPayload1, mesh.DevMsg)
			bindPayload2 := appKeyBind(elemAddr2, appKey.KeyIndex, []byte{0x10, 0x00})
			sendMsg(device.Addr, devKey, bindPayload2, mesh.DevMsg)
		}
		// Add device to group
		group.addDevice(device.Addr)
		// Update net data
		netData = getNetData(netCollection)
		netData.NextAddr = incrementAddr(device.Elements[len(device.Elements)-1].Addr)
		updateNetData(netCollection, netData)
		return device, nil
	})
	obj.FieldFunc("removeDevice", func(args struct{ Addr string }) (Device, error) {
		// Get devKey
		devAddr := decodeBase64(args.Addr)
		device := getDeviceByAddr(devicesCollection, devAddr)
		devKey := getDevKeyByAddr(devKeysCollection, devAddr)
		// Send reset paylode
		resetPaylode := nodeReset()
		sendMsg(devAddr, devKey.Key, resetPaylode, mesh.DevMsg)
		// Remove device from database
		deleteDevice(devicesCollection, devAddr)
		// Remove devkey
		deleteDevKey(devKeysCollection, devAddr)
		// Remove devAddr from group
		group := getGroupByDevAddr(groupsCollection, devAddr)
		group.removeDevice(device.Addr)
		return device, nil
	})
	obj.FieldFunc("removeGroup", func(args struct{ Addr string }) (Group, error) {
		// Get groupAddr
		groupAddr := decodeBase64(args.Addr)
		group := getGroupByAddr(groupsCollection, groupAddr)
		// Delete devices
		for _, devAddr := range group.getDevAddrs() {
			// Get devKey
			devKey := getDevKeyByAddr(devKeysCollection, devAddr)
			// Remove device from database
			deleteDevice(devicesCollection, devAddr)
			// Send reset paylode
			resetPaylode := nodeReset()
			sendMsg(devAddr, devKey.Key, resetPaylode, mesh.DevMsg)
			// Remove devkey
			deleteDevKey(devKeysCollection, devAddr)
		}
		// Remove the group and app key
		deleteGroup(groupsCollection, groupAddr)
		deleteAppKey(appKeysCollection, group.Aid)
		return group, nil
	})
	obj.FieldFunc("addGroup", func(args struct {
		Name string
	}) (Group, error) {
		netData := getNetData(netCollection)
		// Generate an app key
		appKey := make([]byte, 16)
		rand.Read(appKey)
		aid, _ := mesh.GetAid(appKey)
		// Add groups app key
		insertAppKey(appKeysCollection, mesh.AppKey{
			Aid:      []byte{aid},
			Key:      appKey,
			KeyIndex: netData.NetKeyIndex,
		})
		// Add a group
		group := makeGroup(args.Name, netData.NextGroupAddr, []byte{aid})
		// Update net data
		netData.NextGroupAddr = incrementAddr(netData.NextGroupAddr)
		netData.NextAppKeyIndex = incrementKeyIndex(netData.NextAppKeyIndex)
		updateNetData(netCollection, netData)
		return group, nil
	})
	obj.FieldFunc("setState", func(args struct {
		DevAddr    string
		ElemNumber int64
		Value      string
	}) (State, error) {
		value := decodeBase64(args.Value)
		devAddr := decodeBase64(args.DevAddr)
		device := getDeviceByAddr(devicesCollection, devAddr)
		elemNumber := int(args.ElemNumber)
		// Set State
		device.updateState(elemNumber, value)
		// Get appkey from group
		group := getGroupByDevAddr(groupsCollection, device.Addr)
		appKey := getAppKeyByAid(appKeysCollection, group.Aid)
		// Send State
		if device.getState(elemNumber).StateType == "onoff" {
			// Send msg
			onoffPayload := onOffSet(value[0])
			sendMsg(
				device.getElemAddr(elemNumber),
				appKey.Key,
				onoffPayload,
				mesh.AppMsg,
			)
		}
		return device.getState(elemNumber), nil
	})
	obj.FieldFunc("configHub", func() (string, error) {
		// Check if configured
		if getNetData(netCollection).ID != primitive.NilObjectID {
			return "", errors.New("already configured")
		}
		// Stop the mdns server
		mdns.Shutdown()
		// Make a web key
		webKey := make([]byte, 16)
		rand.Read(webKey)
		// Make a net key
		netKey := make([]byte, 16)
		rand.Read(netKey)
		// Clean house
		groupsCollection.DeleteMany(context.TODO(), bson.D{})
		devicesCollection.DeleteMany(context.TODO(), bson.D{})
		appKeysCollection.DeleteMany(context.TODO(), bson.D{})
		devKeysCollection.DeleteMany(context.TODO(), bson.D{})
		netCollection.DeleteMany(context.TODO(), bson.D{})
		// Add and get net data
		insertNetData(netCollection, NetData{
			ID:              primitive.NewObjectID(),
			NetKey:          netKey,
			NetKeyIndex:     []byte{0x00, 0x00},
			NextAppKeyIndex: []byte{0x01, 0x00},
			Flags:           []byte{0x00},
			IvIndex:         []byte{0x00, 0x00, 0x00, 0x01},
			NextAddr:        []byte{0x00, 0x01},
			NextGroupAddr:   []byte{0xc0, 0x00},
			HubSeq:          []byte{0x00, 0x00, 0x00},
			WebKeys:         [][]byte{webKey},
		})
		return encodeBase64(webKey), nil
	})
	obj.FieldFunc("resetHub", func() (bool, error) {
		// Check if configured
		if getNetData(netCollection).ID == primitive.NilObjectID {
			return false, errors.New("not configured")
		}
		// Clean house
		groupsCollection.DeleteMany(context.TODO(), bson.D{})
		devicesCollection.DeleteMany(context.TODO(), bson.D{})
		appKeysCollection.DeleteMany(context.TODO(), bson.D{})
		devKeysCollection.DeleteMany(context.TODO(), bson.D{})
		netCollection.DeleteMany(context.TODO(), bson.D{})
		// Disconnect from proxy
		cln, write, read = nil, nil, nil
		// Mdns
		mdns, _ = zeroconf.Register("hub", "_alexandergherardi._tcp", "local.", 8080, nil, nil)
		return true, nil
	})
}

func registerState(schema *schemabuilder.Schema) {
	obj := schema.Object("State", State{})
	obj.FieldFunc("state", func(ctx context.Context, p *State) string {
		reactive.InvalidateAfter(ctx, 5*time.Second)
		return encodeBase64(p.State)
	})
	obj.FieldFunc("stateType", func(ctx context.Context, p *State) string {
		reactive.InvalidateAfter(ctx, 5*time.Second)
		return p.StateType
	})
}

func registerDevice(schema *schemabuilder.Schema) {
	obj := schema.Object("Device", Device{})
	obj.FieldFunc("type", func(ctx context.Context, p *Device) string {
		reactive.InvalidateAfter(ctx, 5*time.Second)
		return p.Type
	})
	obj.FieldFunc("addr", func(ctx context.Context, p *Device) string {
		reactive.InvalidateAfter(ctx, 5*time.Second)
		return encodeBase64(p.Addr)
	})
	obj.FieldFunc("name", func(ctx context.Context, p *Device) string {
		reactive.InvalidateAfter(ctx, 5*time.Second)
		return p.Name
	})
}

func registerGroup(schema *schemabuilder.Schema) {
	obj := schema.Object("Group", Group{})
	obj.FieldFunc("name", func(ctx context.Context, p *Group) string {
		reactive.InvalidateAfter(ctx, 5*time.Second)
		return p.Name
	})
	obj.FieldFunc("addr", func(ctx context.Context, p *Group) string {
		reactive.InvalidateAfter(ctx, 5*time.Second)
		return encodeBase64(p.Addr)
	})
}

// Schema builds the graphql schema.
func schema() *graphql.Schema {
	builder := schemabuilder.NewSchema()
	registerQuery(builder)
	registerMutation(builder)
	registerDevice(builder)
	registerGroup(builder)
	return builder.MustBuild()
}

// Checks if user is allowed
func authenticate(
	input *graphql.ComputationInput,
	next graphql.MiddlewareNextFunc,
) *graphql.ComputationOutput {
	name := input.ParsedQuery.Selections[0].Name
	// Config hub dose not need a webKey
	if name == "configHub" {
		output := next(input)
		return output
	}
	fmt.Println(input.Variables)
	// Verfiy web key
	netData := getNetData(netCollection)
	if input.Variables["webKey"] == nil {
		return &graphql.ComputationOutput{
			Current: nil,
			Error:   errors.New("Key not found"),
		}
	}
	webKey := decodeBase64(input.Variables["webKey"].(string))
	verify := checkWebKey(netData, webKey)
	if !verify {
		return &graphql.ComputationOutput{
			Current: nil,
			Error:   errors.New("Key invalid"),
		}
	}
	output := next(input)
	return output
}
