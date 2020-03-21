package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	mesh "github.com/AJGherardi/GoMeshCryptro"
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
		return encodeProvData(netData.NetKey, netData.NetKeyIndex, netData.Flags, netData.IvIndex, netData.NextDevAddr)
	})
}

func registerMutation(schema *schemabuilder.Schema) {
	obj := schema.Mutation()
	obj.FieldFunc("addDevice", func(args struct {
		Name    string
		DevType string
		DevKey  string
	}) Device {
		// Message vars
		src := []byte{0x12, 0x34}
		ttl := byte(0x04)
		seq := []byte{0x00, 0x00, 0x00}
		// Decode the dev key
		devKey := decodeBase64(args.DevKey)
		// Generate an app key
		appKey := make([]byte, 16)
		rand.Read(appKey)
		// Send app key add
		addPayload := append([]byte{0x00, 0x00, 0x30, 0x00}, appKey...)
		addMsg, seq := mesh.EncodeAccessMsg(mesh.DevMsg, seq, src, netData.NextDevAddr, ttl, netData.IvIndex, devKey, netKey, addPayload)
		sendProxyPdu(cln, write, addMsg)
		res := <-messages
		fmt.Printf("Res1 %x \n", res)
		// Send app key bind
		bindPayload := []byte{0x80, 0x3d, 0x01, 0x00, 0x03, 0x00, 0x00, 0x10}
		bindMsg, seq := mesh.EncodeAccessMsg(mesh.DevMsg, seq, src, netData.NextDevAddr, ttl, netData.IvIndex, devKey, netKey, bindPayload)
		sendProxyPdu(cln, write, bindMsg)
		res = <-messages
		fmt.Printf("Res2 %x \n", res)
		// Send onoff set
		onoffPayload := []byte{0x82, 0x02, 0x01, 0x00}
		onoffMsg, seq := mesh.EncodeAccessMsg(mesh.AppMsg, seq, src, netData.NextDevAddr, ttl, netData.IvIndex, appKey, netKey, onoffPayload)
		sendProxyPdu(cln, write, onoffMsg)
		res = <-messages
		fmt.Printf("Res3 %x \n", res)
		// Save keys and data
		insertDevKey(devKeysCollection, mesh.DevKey{Addr: netData.NextDevAddr, Key: devKey})
		insertDevice(devicesCollection, Device{Type: args.DevType, Name: args.Name, Addr: netData.NextDevAddr})
		return Device{Type: args.DevType, Name: args.Name}
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
	registerProvData(builder)
	return builder.MustBuild()
}
