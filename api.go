package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	mesh "github.com/AJGherardi/GoMeshCryptro"
	"github.com/AJGherardi/HomeHub/models"
	"github.com/samsarahq/thunder/graphql"
	"github.com/samsarahq/thunder/graphql/schemabuilder"
	"github.com/samsarahq/thunder/reactive"
)

func registerQuery(schema *schemabuilder.Schema) {
	obj := schema.Query()
	obj.FieldFunc("listDevices", func() []Device {
		return getDevices(devicesCollection)
	})
	obj.FieldFunc("getProvData", func() ProvData {
		return encodeProvData(
			netData.NetKey,
			netData.NetKeyIndex,
			netData.Flags,
			netData.IvIndex,
			netData.NextAddr,
		)
	})
}

func registerMutation(schema *schemabuilder.Schema) {
	obj := schema.Mutation()
	obj.FieldFunc("addDevice", func(args struct {
		Name   string
		Addr   string
		DevKey string
	}) Device {
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
		// Send app key bind
		bindPayload := models.AppKeyBind(netData.NextAddr, appKey.KeyIndex, []byte{0x10, 0x00})
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
		bindPayload2 := models.AppKeyBind(netData.NextAddr, appKey.KeyIndex, []byte{0x13, 0x12})
		bindMsg2, seq := mesh.EncodeAccessMsg(
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
		// Get model id
		compPayload := []byte{0x80, 0x50}
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
		// Update net data
		netData.HubSeq = seq
		netData.NextAddr = incrementAddr(netData.NextAddr)
		updateNetData(netCollection, netData)
		// Save Device
		insertDevice(devicesCollection, Device{
			Name: args.Name,
			Addr: netData.NextAddr,
		})
		// Add device to group
		group.DevAddrs = append(group.DevAddrs, netData.NextAddr)
		updateGroup(groupsCollection, group)
		return Device{Name: args.Name}
	})
	obj.FieldFunc("addGroup", func(args struct{ Name string }) Group {
		// Generate an app key
		appKey := make([]byte, 16)
		rand.Read(appKey)
		aid := mesh.GetAid(appKey)
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
}

func registerDevice(schema *schemabuilder.Schema) {
	obj := schema.Object("Device", Device{})
	obj.FieldFunc("type", func(ctx context.Context, p *Device) string {
		reactive.InvalidateAfter(ctx, 5*time.Second)
		return p.Type
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
