package main

import (
	"context"
	"fmt"
	"time"

	mesh "github.com/AJGherardi/GoMeshCryptro"
	"github.com/go-ble/ble"
)

var (
	src = []byte{0x12, 0x34}
	ttl = byte(0x04)
)

func onNotify(req []byte) {
	netData := getNetData(netCollection)
	devKeys := getDevKeys(devicesCollection)
	appKeys := getAppKeys(groupsCollection)
	// Check if it is a network pdu
	if req[0] == 0x00 {
		cmp, err := mesh.DecodePdu(
			msg,
			req[1:],
			netData.getNetKey(),
			netData.getIvIndex(),
			appKeys,
			devKeys,
		)
		if err != nil {
			// New msg if err
			msg = new(mesh.Msg)
			fmt.Println(err)
		}
		// Check if compleat
		if cmp {
			// Sort messages
			switch msg.Payload[0] {
			// Non config models
			case 0x82:
				switch msg.Payload[1] {
				// Onoff Status
				case 0x04:
					onOffStatus(*msg)
				}
				msg = new(mesh.Msg)
			}
		}
	}
	// Check if it is a prov pdu
	if req[0] == 0x03 {
		provMessages <- req[2:]
	}
}

func proxyFilter(a ble.Advertisement) bool {
	if len(a.Services()) > 0 {
		service := a.Services()[0]
		return ble.UUID16(0x1828).Equal(service)
	}
	return false
}

func provFilter(a ble.Advertisement) bool {
	if len(a.Services()) > 0 {
		service := a.Services()[0]
		return ble.UUID16(0x1827).Equal(service)
	}
	return false
}

func reconnectOnDisconnect(ch <-chan struct{}) {
	// Wait for close
	<-ch
	// Dont reconect if there are no devices
	if len(getDevices(devicesCollection)) == 0 {
		cln, write, read = nil, nil, nil
		return
	}
	// Reconnect to proxy
	cln, write, read = connectToProxy()
	// Rerun reconnectOnDisconnect
	go reconnectOnDisconnect(cln.Disconnected())
}

func connectToProxy() (ble.Client, *ble.Characteristic, *ble.Characteristic) {
	// Find and Connect to Mesh Node
	ctx := context.TODO()
	cln, err := ble.Connect(ctx, proxyFilter)
	if err != nil {
		fmt.Println("Connect err " + err.Error())
	}
	// Set Mtu
	cln.ExchangeMTU(128)
	// Get Characteristics from Profile
	p, _ := cln.DiscoverProfile(true)
	write := p.FindCharacteristic(ble.NewCharacteristic(ble.UUID16(0x2add)))
	read := p.FindCharacteristic(ble.NewCharacteristic(ble.UUID16(0x2ade)))
	// Subscribe to mesh Out Characteristic
	cln.Subscribe(read, false, onNotify)
	return cln, write, read
}

func findDevices() []string {
	ctx := ble.WithSigHandler(context.WithTimeout(context.Background(), 5*time.Second))
	adv, _ := ble.Find(ctx, false, provFilter)
	var addrs []string
	for _, devices := range adv {
		addrs = append(addrs, devices.Addr().String())
	}
	return addrs
}

func connectToUnprovisioned(addr string) (ble.Client, *ble.Characteristic, *ble.Characteristic) {
	// Find and Connect to Mesh Node
	ctx := context.TODO()
	cln, err := ble.Connect(
		ctx,
		func(a ble.Advertisement) bool {
			if len(a.Services()) > 0 {
				devAddr := a.Addr().String()
				return addr == devAddr
			}
			return false
		},
	)
	if err != nil {
		fmt.Println("Connect err " + err.Error())
	}
	// Set Mtu
	cln.ExchangeMTU(128)
	// Get Characteristics from Profile
	p, _ := cln.DiscoverProfile(true)
	write := p.FindCharacteristic(ble.NewCharacteristic(ble.UUID16(0x2adb)))
	read := p.FindCharacteristic(ble.NewCharacteristic(ble.UUID16(0x2adc)))
	// Subscribe to mesh Out Characteristic
	cln.Subscribe(read, false, onNotify)
	return cln, write, read
}

func sendProxyPdu(cln ble.Client, write *ble.Characteristic, msg [][]byte) {
	for _, pdu := range msg {
		proxyPdu := append([]byte{0x00}, pdu...)
		err := cln.WriteCharacteristic(write, proxyPdu, true)
		if err != nil {
			fmt.Println("Send err " + err.Error())
		}
	}
}

func sendProvPdu(cln ble.Client, write *ble.Characteristic, pdu []byte) {
	proxyPdu := append([]byte{0x03}, pdu...)
	err := cln.WriteCharacteristic(write, proxyPdu, true)
	if err != nil {
		fmt.Println("Send prov err " + err.Error())
	}
}

// Sends a mesh message and returns a response
func sendMsg(dst []byte, key []byte, payload []byte, msgType mesh.MsgType) {
	netData := getNetData(netCollection)
	// Encode msg and get new seq
	msg, seq, err := mesh.EncodeAccessMsg(
		msgType,
		netData.getHubSeq(),
		src,
		dst,
		ttl,
		netData.getIvIndex(),
		key,
		netData.getNetKey(),
		payload,
	)
	if err != nil {
		fmt.Println("Encode err " + err.Error())
	}
	// Send msg
	sendProxyPdu(cln, write, msg)
	// Update seq
	netData.updateHubSeq(seq)
}

func provisionDevice(cln ble.Client, write *ble.Characteristic, netKey, keyIndex, flags, ivIndex, devAddr []byte) []byte {
	// Gen prov keys
	provPrivKey, provPubKeyX, provPubKeyY, _ := mesh.GenerateProvKeys()
	provPubKey := append(provPubKeyX, provPubKeyY...)
	// Send invite
	invite := []byte{0x00, 0x00}
	sendProvPdu(cln, write, invite)
	// Recive Capabilities
	capabilities := <-provMessages
	// Send start
	start := []byte{0x00, 0x00, 0x00, 0x00, 0x00}
	startPdu := append([]byte{0x02}, start...)
	sendProvPdu(cln, write, startPdu)
	// Send public Key
	pubKeyPdu := append([]byte{0x03}, provPubKey...)
	sendProvPdu(cln, write, pubKeyPdu)
	// Recive dev public key
	devPubKey := <-provMessages
	// Calculate shared secret
	secret := mesh.GetSharedSecret(devPubKey, provPrivKey)
	// Get inputs
	inputs := append([]byte{0x00}, capabilities...)
	inputs = append(inputs, []byte{0x00, 0x00, 0x00, 0x00, 0x00}...)
	inputs = append(inputs, provPubKey...)
	inputs = append(inputs, devPubKey...)
	// Generate confirmation
	conf, provRandom, confSalt, _ := mesh.GenerateConfData(inputs, secret)
	// Send confirmation
	confPdu := append([]byte{0x05}, conf...)
	sendProvPdu(cln, write, confPdu)
	// Recive confirmation
	_ = <-provMessages
	// Send random
	randomPdu := append([]byte{0x06}, provRandom...)
	sendProvPdu(cln, write, randomPdu)
	// Recive random
	devRandom := <-provMessages
	provData := append(netKey, keyIndex...)
	provData = append(provData, flags...)
	provData = append(provData, ivIndex...)
	provData = append(provData, devAddr...)
	fmt.Printf("provData %x \n", provData)
	// Gen and send prov data
	encProvData, provSalt, _ := mesh.GetProvData(secret, confSalt, provRandom, devRandom, netKey, keyIndex, flags, ivIndex, devAddr)
	// Send prov data
	provDataPdu := append([]byte{0x07}, encProvData...)
	sendProvPdu(cln, write, provDataPdu)
	// Recive Complete
	_ = <-provMessages
	// Get Dev Key
	devKey, _ := mesh.GetDevKey(secret, provSalt)
	return devKey
}
