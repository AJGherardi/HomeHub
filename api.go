package main

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"reflect"
	"time"

	mesh "github.com/AJGherardi/GoMeshCryptro"
	"github.com/AJGherardi/HomeHub/models"
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
	obj.FieldFunc("getProvData", func() (ProvData, error) {
		netData := getNetData(netCollection)
		return encodeProvData(
			netData.NetKey,
			netData.NetKeyIndex,
			netData.Flags,
			netData.IvIndex,
			netData.NextAddr,
		), nil
	})
	obj.FieldFunc("getState", func(args struct {
		DevAddr    string
		ElemNumber int64
	}) (State, error) {
		devAddr := decodeBase64(args.DevAddr)
		device := getDeviceByAddr(devicesCollection, devAddr)
		element := device.Elements[args.ElemNumber]
		return element.State, nil
	})
}

func registerMutation(schema *schemabuilder.Schema) {
	obj := schema.Mutation()
	obj.FieldFunc("addDevice", func(args struct {
		Name   string
		Addr   string
		DevKey string
	}) (Device, error) {
		netData := getNetData(netCollection)
		// Create device object
		var device Device
		device.Name = args.Name
		device.Addr = netData.NextAddr
		// Get group
		groupAddr := decodeBase64(args.Addr)
		group := getGroupByAddr(groupsCollection, groupAddr)
		// Get app key
		appKey := getAppKeyByAid(appKeysCollection, group.Aid)
		// Decode the dev key
		devKey := decodeBase64(args.DevKey)
		insertDevKey(devKeysCollection, mesh.DevKey{Addr: device.Addr, Key: devKey})
		// Send app key add
		addPayload := models.AppKeyAdd(netData.NetKeyIndex, appKey.KeyIndex, appKey.Key)
		addRsp := sendMsgWithRsp(device.Addr, devKey, addPayload, mesh.DevMsg)
		fmt.Printf("add %x \n", addRsp)
		// Send app key bind for config data
		bindPayload := models.AppKeyBind(device.Addr, appKey.KeyIndex, []byte{0x13, 0x12})
		bindRsp := sendMsgWithRsp(device.Addr, devKey, bindPayload, mesh.DevMsg)
		fmt.Printf("bind %x \n", bindRsp)
		// Get model id
		compPayload := models.ConfigDataGet()
		compRsp := sendMsgWithRsp(device.Addr, appKey.Key, compPayload, mesh.AppMsg)
		fmt.Printf("comp %x \n", compRsp)
		if reflect.DeepEqual(compRsp[2:], []byte{0x00, 0x00}) {
			devType := "2PowerSwitch"
			elemAddr1 := device.Addr
			elemAddr2 := incrementAddr(elemAddr1)
			// Send app key bind for onoff
			bindPayload1 := models.AppKeyBind(elemAddr1, appKey.KeyIndex, []byte{0x10, 0x00})
			bindRsp1 := sendMsgWithRsp(device.Addr, devKey, bindPayload1, mesh.DevMsg)
			fmt.Printf("bind %x \n", bindRsp1)
			bindPayload2 := models.AppKeyBind(elemAddr2, appKey.KeyIndex, []byte{0x10, 0x00})
			bindRsp2 := sendMsgWithRsp(device.Addr, devKey, bindPayload2, mesh.DevMsg)
			fmt.Printf("bind %x \n", bindRsp2)
			// Set type and elements
			device.Type = devType
			device.Elements = []Element{
				Element{Addr: elemAddr1, State: State{
					StateType: "onoff",
					State:     []byte{0x00},
				}},
				Element{Addr: elemAddr2, State: State{
					StateType: "onoff",
					State:     []byte{0x00},
				}},
			}
		}
		// Save Device
		insertDevice(devicesCollection, device)
		// Add device to group
		group.DevAddrs = append(group.DevAddrs, device.Addr)
		updateGroup(groupsCollection, group)
		// Update net data
		netData = getNetData(netCollection)
		netData.NextAddr = incrementAddr(device.Elements[len(device.Elements)-1].Addr)
		updateNetData(netCollection, netData)
		return device, nil
	})
	obj.FieldFunc("addGroup", func(args struct {
		Name string
	}) (Group, error) {
		var group Group
		netData := getNetData(netCollection)
		// Generate an app key
		appKey := make([]byte, 16)
		rand.Read(appKey)
		aid := mesh.GetAid(appKey)
		// Add groups app key
		insertAppKey(appKeysCollection, mesh.AppKey{
			Aid:      []byte{aid},
			Key:      appKey,
			KeyIndex: netData.NetKeyIndex,
		})
		// Add a group
		group.Name = args.Name
		group.Addr = netData.NextGroupAddr
		group.Aid = []byte{aid}
		insertGroup(groupsCollection, group)
		// Update net data
		netData.NextGroupAddr = incrementAddr(netData.NextGroupAddr)
		netData.NextAppKeyIndex = models.IncrementKeyIndex(netData.NextAppKeyIndex)
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
		// Set State
		device.Elements[args.ElemNumber].State.State = value
		// Get appkey from group
		group := getGroupByDevAddr(groupsCollection, device.Addr)
		appKey := getAppKeyByAid(appKeysCollection, group.Aid)
		// Send State
		if device.Elements[args.ElemNumber].State.StateType == "onoff" {
			// Send msg
			onoffPayload := models.OnOffSet(value[0])
			sendMsgWithRsp(
				device.Elements[args.ElemNumber].Addr,
				appKey.Key,
				onoffPayload,
				mesh.AppMsg,
			)
		}
		updateDevice(devicesCollection, device)
		return device.Elements[args.ElemNumber].State, nil
	})
	obj.FieldFunc("configHub", func() string {
		// Make a web key
		webKey := make([]byte, 16)
		rand.Read(webKey)
		// Clean house
		groupsCollection.DeleteMany(context.TODO(), bson.D{})
		devicesCollection.DeleteMany(context.TODO(), bson.D{})
		webKeysCollection.DeleteMany(context.TODO(), bson.D{})
		appKeysCollection.DeleteMany(context.TODO(), bson.D{})
		devKeysCollection.DeleteMany(context.TODO(), bson.D{})
		netCollection.DeleteMany(context.TODO(), bson.D{})
		// Add and get net data
		insertNetData(netCollection, NetData{
			ID: primitive.NewObjectID(),
			NetKey: []byte{0xaf, 0xc3, 0x27, 0x0e, 0xda,
				0x88, 0x02, 0xf7, 0x2c, 0x1e, 0x53,
				0x24, 0x38, 0xa9, 0x79, 0xeb,
			},
			NetKeyIndex:     []byte{0x00, 0x00},
			NextAppKeyIndex: []byte{0x01, 0x00},
			Flags:           []byte{0x00},
			IvIndex:         []byte{0x00, 0x00, 0x00, 0x00},
			NextAddr:        []byte{0x00, 0x01},
			NextGroupAddr:   []byte{0xc0, 0x00},
			HubSeq:          []byte{0x00, 0x00, 0x00},
			WebKeys:         [][]byte{webKey},
		})
		return encodeBase64(webKey)
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

func registerProvData(schema *schemabuilder.Schema) {
	obj := schema.Object("ProvData", ProvData{})
	obj.FieldFunc("networkKey", func(ctx context.Context, p *ProvData) string {
		reactive.InvalidateAfter(ctx, 5*time.Second)
		return p.NetworkKey
	})
	obj.FieldFunc("keyIndex", func(ctx context.Context, p *ProvData) string {
		reactive.InvalidateAfter(ctx, 5*time.Second)
		return p.KeyIndex
	})
	obj.FieldFunc("flags", func(ctx context.Context, p *ProvData) string {
		reactive.InvalidateAfter(ctx, 5*time.Second)
		return p.Flags
	})
	obj.FieldFunc("ivIndex", func(ctx context.Context, p *ProvData) string {
		reactive.InvalidateAfter(ctx, 5*time.Second)
		return p.IvIndex
	})
	obj.FieldFunc("nextDevAddr", func(ctx context.Context, p *ProvData) string {
		reactive.InvalidateAfter(ctx, 5*time.Second)
		return p.NextDevAddr
	})
}

// Schema builds the graphql schema.
func schema() *graphql.Schema {
	builder := schemabuilder.NewSchema()
	registerQuery(builder)
	registerMutation(builder)
	registerDevice(builder)
	registerGroup(builder)
	registerProvData(builder)
	return builder.MustBuild()
}

// Checks if user is allowed
func authenticate(input *graphql.ComputationInput, next graphql.MiddlewareNextFunc) *graphql.ComputationOutput {
	name := input.ParsedQuery.Selections[0].Name
	// Config hub dose not need a webKey
	if name == "configHub" {
		output := next(input)
		return output
	}
	// Verfiy web key
	netData := getNetData(netCollection)
	webKey := decodeBase64(input.Variables["webKey"].(string))
	verify := checkWebKey(netData, webKey)
	if !verify {
		return &graphql.ComputationOutput{Current: nil, Error: errors.New("Key not found")}
	}
	output := next(input)
	return output
}
