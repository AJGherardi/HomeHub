package main

import (
	"context"
	"fmt"
	"time"

	mesh "github.com/AJGherardi/GoMeshCryptro"
	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
)

func onNotify(req []byte) {
	devKeys := getDevKeys(devKeysCollection)
	appKeys := getAppKeys(appKeysCollection)
	if req[0] == 0x00 {
		msg := new(mesh.Msg)
		mesh.DecodePdu(msg, req[1:], netData.NetKey, netData.IvIndex, appKeys, devKeys)
		if len(msg.Payload) != 0 {
			messages <- msg.Payload
		}
	}
}

func filter(a ble.Advertisement) bool {
	if len(a.Services()) > 0 {
		service := a.Services()[0]
		return ble.UUID16(0x1828).Equal(service)
	}
	return false
}

func connectToProxy() (ble.Client, *ble.Characteristic) {
	d, err := dev.NewDevice("default")
	if err != nil {
		fmt.Println(err.Error())
	}
	ble.SetDefaultDevice(d)
	// Find and Connect to Mesh Node
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), 15*time.Second))
	cln, _ := ble.Connect(ctx, filter)
	// Set Mtu
	cln.ExchangeMTU(128)
	// Get Characteristics from Profile
	p, _ := cln.DiscoverProfile(true)
	write := p.FindCharacteristic(ble.NewCharacteristic(ble.UUID16(0x2add)))
	read := p.FindCharacteristic(ble.NewCharacteristic(ble.UUID16(0x2ade)))
	// Subscribe to mesh Out Characteristic
	cln.Subscribe(read, false, onNotify)
	return cln, write
}

func sendProxyPdu(cln ble.Client, write *ble.Characteristic, msg [][]byte) {
	for _, pdu := range msg {
		proxyPdu := append([]byte{0x00}, pdu...)
		cln.WriteCharacteristic(write, proxyPdu, false)
	}
}
