package main

import (
	"context"
	"fmt"
	"time"

	mesh "github.com/AJGherardi/GoMeshCryptro"
	"github.com/go-ble/ble"
)

var (
	src          = []byte{0x12, 0x34}
	ttl          = byte(0x04)
	messages     = make(map[[2]byte](chan []byte))
	provMessages = make(chan []byte)
	msg          = new(mesh.Msg)
	d            ble.Device
)

func onNotify(req []byte) {
	netData := getNetData(netCollection)
	devKeys := getDevKeys(devKeysCollection)
	appKeys := getAppKeys(appKeysCollection)
	// Check if it is a network pdu
	if req[0] == 0x00 {
		cmp, err := mesh.DecodePdu(
			msg,
			req[1:],
			netData.NetKey,
			netData.IvIndex,
			appKeys,
			devKeys,
		)
		if err != nil {
			fmt.Println(err)
		}
		if cmp {
			var msgSrc [2]byte
			copy(msgSrc[:], msg.Src)
			messages[msgSrc] <- msg.Payload
			msg = new(mesh.Msg)
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
	// Check if open
	_, open := <-ch
	// Dont reconect if there are no devices
	if len(getDevices(devicesCollection)) == 0 {
		cln, write = nil, nil
		return
	}
	if !open {
		messages = make(map[[2]byte](chan []byte))
		cln, write, read = connectToProxy()
		go reconnectOnDisconnect(cln.Disconnected())
	}
}

func connectToProxy() (ble.Client, *ble.Characteristic, *ble.Characteristic) {
	// Find and Connect to Mesh Node
	ctx := context.TODO()
	cln, err := ble.Connect(ctx, proxyFilter)
	if err != nil {
		fmt.Println(err)
	}
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
		fmt.Println(err)
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
		err := cln.WriteCharacteristic(write, proxyPdu, false)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
}

func sendProvPdu(cln ble.Client, write *ble.Characteristic, pdu []byte) {
	proxyPdu := append([]byte{0x03}, pdu...)
	cln.WriteCharacteristic(write, proxyPdu, false)
}

// Sends a mesh message without a response
func sendMsgWithoutRsp(dst []byte, key []byte, payload []byte, msgType mesh.MsgType) {
	netData := getNetData(netCollection)
	// Encode msg and get new seq
	msg, seq, err := mesh.EncodeAccessMsg(
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
	if err != nil {
		fmt.Println(err)
	}
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
	msg, seq, err := mesh.EncodeAccessMsg(
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
	if err != nil {
		fmt.Println(err.Error())
	}
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
