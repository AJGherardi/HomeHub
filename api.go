package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/samsarahq/thunder/graphql"
	"github.com/samsarahq/thunder/graphql/schemabuilder"
	"github.com/samsarahq/thunder/reactive"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ControlResponse is the response for a listControl query
type ControlResponse struct {
	Devices []Device
	Groups  []Group
}

func registerQuery(schema *schemabuilder.Schema) {
	obj := schema.Query()
	obj.FieldFunc("listControl", func() (ControlResponse, error) {
		rsp := ControlResponse{
			Groups:  getGroups(groupsCollection),
			Devices: getDevices(devicesCollection),
		}
		return rsp, nil
	})
	obj.FieldFunc("availableDevices", func() ([]string, error) {
		uuids := make([]string, 0)
		for _, uuid := range unprovisionedNodes {
			b64 := encodeBase64(uuid)
			uuids = append(uuids, b64)
		}
		return uuids, nil
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
		DevUUID string
	}) (Device, error) {
		// Provision device
		uuid := decodeBase64(args.DevUUID)
		controller.Provision(uuid)
		// Wait for node added
		addr := <-nodeAdded
		// Create device object
		device := makeDevice(
			args.Name,
			"2PowerSwitch",
			addr,
		)
		// Get group
		groupAddr := decodeBase64(args.Addr)
		group := getGroupByAddr(groupsCollection, groupAddr)
		// Get model id
		if true {
			// Set type and add elements
			device.addElem("onoff")
			device.addElem("onoff")
		}
		// Configure Device
		controller.ConfigureNode(device.Addr, group.KeyIndex)
		// Add device to group
		group.addDevice(device.Addr)
		return device, nil
	})
	obj.FieldFunc("removeDevice", func(args struct{ Addr string }) (Device, error) {
		// Get devKey
		devAddr := decodeBase64(args.Addr)
		device := getDeviceByAddr(devicesCollection, devAddr)
		// Send reset paylode
		controller.ResetNode(devAddr)
		// Remove device from database
		deleteDevice(devicesCollection, devAddr)
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
			device := getDeviceByAddr(devicesCollection, devAddr)
			// Send reset paylode
			controller.ResetNode(device.Addr)
			// Remove device from database
			deleteDevice(devicesCollection, devAddr)
		}
		// Remove the group
		deleteGroup(groupsCollection, groupAddr)
		return group, nil
	})
	obj.FieldFunc("addGroup", func(args struct {
		Name string
	}) (Group, error) {
		netData := getNetData(netCollection)
		// Get net values
		keyIndex := netData.getNextAppKeyIndex()
		groupAddr := netData.getNextGroupAddr()
		// Add an app key
		controller.AddKey(keyIndex)
		// Add a group
		group := makeGroup(args.Name, groupAddr, keyIndex)
		// Update net data
		netData.incrementNextGroupAddr()
		netData.incrementNextAppKeyIndex()
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
		// Send State
		if device.getState(elemNumber).StateType == "onoff" {
			// Send msg
			controller.SendMessage(
				value[0],
				device.getElemAddr(elemNumber),
				group.KeyIndex,
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
		// Clean house
		groupsCollection.DeleteMany(context.TODO(), bson.D{})
		devicesCollection.DeleteMany(context.TODO(), bson.D{})
		netCollection.DeleteMany(context.TODO(), bson.D{})
		// Add and get net data
		makeNetData(webKey)
		// Setup controller
		controller.Setup()
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
		netCollection.DeleteMany(context.TODO(), bson.D{})
		// Reset mesh controller
		controller.Reset()
		time.Sleep(time.Second)
		controller.Reboot()
		// Start Mdns
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
	obj.FieldFunc("devAddrs", func(ctx context.Context, p *Group) []string {
		reactive.InvalidateAfter(ctx, 5*time.Second)
		var result []string
		for _, devAddr := range p.DevAddrs {
			devAddrString := encodeBase64(devAddr)
			result = append(result, devAddrString)
		}
		return result
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
	verify := netData.checkWebKey(webKey)
	if !verify {
		return &graphql.ComputationOutput{
			Current: nil,
			Error:   errors.New("Key invalid"),
		}
	}
	output := next(input)
	return output
}
