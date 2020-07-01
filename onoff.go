package main

import (
	"fmt"

	mesh "github.com/AJGherardi/GoMeshCryptro"
)

// OnOffGet makes an generic onoff get payload
func onOffGet() []byte {
	opcode := []byte{0x82, 0x01}
	return opcode
}

// OnOffSet makes an generic onoff set payload
func onOffSet(onoff byte) []byte {
	opcode := []byte{0x82, 0x02}
	output := append(opcode, []byte{onoff, 0x00}...)
	return output
}

// OnOffStatus handles an onoff status msg
func onOffStatus(msg mesh.Msg) {
	device := getDeviceByElemAddr(devicesCollection, msg.Src)
	state := msg.Payload[2]
	fmt.Printf("status %x \n", msg.Payload)
	// Update the state
	device.updateStateUsingAddr(msg.Src, []byte{state})
}
