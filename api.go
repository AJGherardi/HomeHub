package main

import (
	"context"
	"crypto/rand"
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
	obj.FieldFunc("listDevices", func(args struct{ WebKey string }) []Device {
		// Verify webKey
		webKey := decodeBase64(args.WebKey)
		verify := checkWebKey(netData, webKey)
		if !verify {
			return []Device{}
		}
		return getDevices(devicesCollection)
	})
	obj.FieldFunc("getProvData", func(args struct{ WebKey string }) ProvData {
		// Verify webKey
		webKey := decodeBase64(args.WebKey)
		verify := checkWebKey(netData, webKey)
		if !verify {
			return ProvData{}
		}
		return encodeProvData(
			netData.NetKey,
			netData.NetKeyIndex,
			netData.Flags,
			netData.IvIndex,
			netData.NextAddr,
		)
	})
	obj.FieldFunc("getState", func(args struct {
		DevAddr    string
		ElemNumber int64
		WebKey     string
	}) State {
		// Verify webKey
		webKey := decodeBase64(args.WebKey)
		verify := checkWebKey(netData, webKey)
		if !verify {
			return State{}
		}
		devAddr := decodeBase64(args.DevAddr)
		device := getDeviceByAddr(devicesCollection, devAddr)
		element := device.Elements[args.ElemNumber]
		return element.State
	})
}

func registerMutation(schema *schemabuilder.Schema) {
	obj := schema.Mutation()
	obj.FieldFunc("addDevice", func(args struct {
		Name   string
		Addr   string
		DevKey string
		WebKey string
	}) Device {
		// Verify webKey
		webKey := decodeBase64(args.WebKey)
		verify := checkWebKey(netData, webKey)
		if !verify {
			return Device{}
		}
		// Message vars
		src := []byte{0x12, 0x34}
		ttl := byte(0x04)
		seq := netData.HubSeq
		// Get group
		groupAddr := decodeBase64(args.Addr)
		group := getGroupByAddr(groupsCollection, groupAddr)
		// Get app key
		appKey := getAppKeyByAid(appKeysCollection, group.Aid)
		// Decode the dev key
		devKey := decodeBase64(args.DevKey)
		insertDevKey(devKeysCollection, mesh.DevKey{Addr: netData.NextAddr, Key: devKey})
		// Send app key add
		addPayload := models.AppKeyAdd(netData.NetKeyIndex, appKey.KeyIndex, appKey.Key)
		addMsg, seq := mesh.EncodeAccessMsg(
			mesh.DevMsg,
			seq,
			src,
			netData.NextAddr,
			ttl,
			netData.IvIndex,
			devKey,
			netData.NetKey,
			addPayload,
		)
		sendProxyPdu(cln, write, addMsg)
		res := <-messages
		fmt.Printf("add %x \n", res)
		// Send app key bind for config data
		bindPayload := models.AppKeyBind(netData.NextAddr, appKey.KeyIndex, []byte{0x13, 0x12})
		bindMsg, seq := mesh.EncodeAccessMsg(
			mesh.DevMsg,
			seq,
			src,
			netData.NextAddr,
			ttl,
			netData.IvIndex,
			devKey,
			netData.NetKey,
			bindPayload,
		)
		sendProxyPdu(cln, write, bindMsg)
		res = <-messages
		fmt.Printf("bind %x \n", res)
		// Get model id
		compPayload := models.ConfigDataGet()
		compMsg, seq := mesh.EncodeAccessMsg(
			mesh.AppMsg,
			seq,
			src,
			netData.NextAddr,
			ttl,
			netData.IvIndex,
			appKey.Key,
			netData.NetKey,
			compPayload,
		)
		sendProxyPdu(cln, write, compMsg)
		res = <-messages
		fmt.Printf("comp %x \n", res)
		var device Device
		var elemAddr1 []byte = netData.NextAddr
		var elemAddr2 []byte = incrementAddr(elemAddr1)
		if reflect.DeepEqual(res[2:], []byte{0x00, 0x00}) {
			devType := "2PowerSwitch"
			// Send app key bind for onoff
			var bindMsg1 [][]byte
			bindPayload1 := models.AppKeyBind(elemAddr1, appKey.KeyIndex, []byte{0x10, 0x00})
			bindMsg1, seq = mesh.EncodeAccessMsg(
				mesh.DevMsg,
				seq,
				src,
				netData.NextAddr,
				ttl,
				netData.IvIndex,
				devKey,
				netData.NetKey,
				bindPayload1,
			)
			sendProxyPdu(cln, write, bindMsg1)
			res = <-messages
			fmt.Printf("bind %x \n", res)
			var bindMsg2 [][]byte
			bindPayload2 := models.AppKeyBind(elemAddr2, appKey.KeyIndex, []byte{0x10, 0x00})
			bindMsg2, seq = mesh.EncodeAccessMsg(
				mesh.DevMsg,
				seq,
				src,
				netData.NextAddr,
				ttl,
				netData.IvIndex,
				devKey,
				netData.NetKey,
				bindPayload2,
			)
			sendProxyPdu(cln, write, bindMsg2)
			res = <-messages
			fmt.Printf("bind %x \n", res)
			// Make Device
			device = Device{
				Name: args.Name,
				Addr: netData.NextAddr,
				Type: devType,
				Elements: []Element{
					Element{Addr: elemAddr1, State: State{
						StateType: "onoff",
						State:     []byte{0x00},
					}},
					Element{Addr: elemAddr2, State: State{
						StateType: "onoff",
						State:     []byte{0x00},
					}},
				},
			}
		}
		// Save Device
		insertDevice(devicesCollection, device)
		// Add device to group
		group.DevAddrs = append(group.DevAddrs, netData.NextAddr)
		updateGroup(groupsCollection, group)
		// Update net data
		netData.HubSeq = seq
		netData.NextAddr = incrementAddr(elemAddr2)
		updateNetData(netCollection, netData)
		return device
	})
	obj.FieldFunc("addGroup", func(args struct {
		Name   string
		WebKey string
	}) Group {
		// Verify webKey
		webKey := decodeBase64(args.WebKey)
		verify := checkWebKey(netData, webKey)
		if !verify {
			return Group{}
		}
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
		insertGroup(groupsCollection, Group{
			Name: args.Name,
			Addr: netData.NextGroupAddr,
			Aid:  []byte{aid}},
		)
		// Update net data
		netData.NextGroupAddr = incrementAddr(netData.NextGroupAddr)
		netData.NextAppKeyIndex = models.IncrementKeyIndex(netData.NextAppKeyIndex)
		updateNetData(netCollection, netData)
		return Group{Name: args.Name, Addr: netData.NextGroupAddr}
	})
	obj.FieldFunc("setState", func(args struct {
		DevAddr    string
		ElemNumber int64
		Value      string
		WebKey     string
	}) State {
		// Verify webKey
		webKey := decodeBase64(args.WebKey)
		verify := checkWebKey(netData, webKey)
		if !verify {
			return State{}
		}
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
			// Message vars
			src := []byte{0x12, 0x34}
			ttl := byte(0x04)
			seq := netData.HubSeq
			// Send msg
			onoffPayload := models.OnOffSet(value[0])
			onoffMsg, seq := mesh.EncodeAccessMsg(
				mesh.AppMsg,
				seq,
				src,
				device.Elements[args.ElemNumber].Addr,
				ttl,
				netData.IvIndex,
				appKey.Key,
				netData.NetKey,
				onoffPayload,
			)
			sendProxyPdu(cln, write, onoffMsg)
			res := <-messages
			fmt.Printf("onoff %x \n", res)
			// Update net data
			netData.HubSeq = seq
			updateNetData(netCollection, netData)
		}
		updateDevice(devicesCollection, device)
		return device.Elements[args.ElemNumber].State
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
		netData = getNetData(netCollection)
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
