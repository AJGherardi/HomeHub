package main

import (
	"context"
	"fmt"
	"time"

	mesh "github.com/AJGherardi/GoMeshCryptro"
	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
)

var (
	src      = []byte{0x12, 0x34}
	ttl      = byte(0x04)
	messages = make(map[[2]byte](chan []byte))
	msg      = new(mesh.Msg)
	d        ble.Device
)

func onNotify(req []byte) {
	netData := getNetData(netCollection)
	devKeys := getDevKeys(devKeysCollection)
	appKeys := getAppKeys(appKeysCollection)
	// Check if it is a network pdu
	if req[0] == 0x00 {
		cmp := mesh.DecodePdu(msg, req[1:], netData.NetKey, netData.IvIndex, appKeys, devKeys)
		if cmp {
			var msgSrc [2]byte
			copy(msgSrc[:], msg.Src)
			messages[msgSrc] <- msg.Payload
			msg = new(mesh.Msg)
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

func reconnectOnDisconnect(ch <-chan struct{}) {
	// Check if open
	_, open := <-ch
	if len(getDevices(devicesCollection)) == 0 {
		cln, write = nil, nil
		return
	}
	if !open {
		messages = make(map[[2]byte](chan []byte))
		cln, write = connectToProxy()
		go reconnectOnDisconnect(cln.Disconnected())
	}
}

func connectToProxy() (ble.Client, *ble.Characteristic) {
	if d == nil {
		d, _ = dev.NewDevice("default")
		ble.SetDefaultDevice(d)
	}

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
	// Set up receivers
	devices := getDevices(devicesCollection)
	for _, device := range devices {
		for _, element := range device.Elements {
			addReceiver(element.Addr)
		}
	}
	fmt.Println("con")
	return cln, write
}

func sendProxyPdu(cln ble.Client, write *ble.Characteristic, msg [][]byte) {
	for _, pdu := range msg {
		proxyPdu := append([]byte{0x00}, pdu...)
		cln.WriteCharacteristic(write, proxyPdu, false)
	}
}

// Sends a mesh message without a response
func sendMsgWithoutRsp(dst []byte, key []byte, payload []byte, msgType mesh.MsgType) {
	netData := getNetData(netCollection)
	// Encode msg and get new seq
	msg, seq := mesh.EncodeAccessMsg(
		msgType,
		netData.HubSeq,
		src,
		dst,
		ttl,
		netData.IvIndex,
		key,
		netData.NetKey,
		payload,
	)
	// Send msg
	sendProxyPdu(cln, write, msg)
	// Update seq
	netData.HubSeq = seq
	updateNetData(netCollection, netData)
}

// Sends a mesh message and returns a response
func sendMsgWithRsp(dst []byte, key []byte, payload []byte, msgType mesh.MsgType) []byte {
	netData := getNetData(netCollection)
	// Encode msg and get new seq
	msg, seq := mesh.EncodeAccessMsg(
		msgType,
		netData.HubSeq,
		src,
		dst,
		ttl,
		netData.IvIndex,
		key,
		netData.NetKey,
		payload,
	)
	// Send msg
	sendProxyPdu(cln, write, msg)
	// Update seq
	netData.HubSeq = seq
	updateNetData(netCollection, netData)
	// Get Rsp from addr
	var msgDst [2]byte
	copy(msgDst[:], dst)
	res := <-messages[msgDst]
	return res
}

// Adds a receiver for one address
func addReceiver(addr []byte) {
	var recAddr [2]byte
	copy(recAddr[:], addr)
	messages[recAddr] = make(chan []byte)
}
