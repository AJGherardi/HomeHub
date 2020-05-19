package models

import (
	"encoding/binary"
)

// NodeReset resets a node
func NodeReset() []byte {
	opcode := []byte{0x80, 0x49}
	return opcode
}

// AppKeyAdd makes an appkey add payload
func AppKeyAdd(netIndex []byte, appIndex []byte, appKey []byte) []byte {
	opcode := []byte{0x00}
	indices := encodeKeyIndices(netIndex, appIndex)
	payload := append(opcode, indices...)
	payload = append(payload, appKey...)
	return payload
}

// AppKeyBind makes an appkey bind payload
func AppKeyBind(addr []byte, appIndex []byte, modelID []byte) []byte {
	opcode := []byte{0x80, 0x3d}
	elemAddr := []byte{addr[1], addr[0]}
	model := []byte{modelID[1], modelID[0]}
	payload := append(opcode, elemAddr...)
	payload = append(payload, appIndex...)
	payload = append(payload, model...)
	return payload
}

// ConfigDataGet makes an config data get payload
func ConfigDataGet() []byte {
	opcode := []byte{0x80, 0x50}
	return opcode
}

// IncrementKeyIndex increments an key index
func IncrementKeyIndex(imput []byte) []byte {
	// Convert to uint16
	index := binary.LittleEndian.Uint16(imput)
	// Increment
	index++
	// Convert back to byte slice
	output := make([]byte, 2)
	binary.LittleEndian.PutUint16(output, index)
	return output
}

func encodeKeyIndices(netIndex []byte, appIndex []byte) []byte {
	// Remove Padding
	netInt := binary.LittleEndian.Uint16(netIndex)
	appInt := binary.LittleEndian.Uint16(appIndex)
	// Add indeces
	indices := uint32(appInt)<<12 | uint32(netInt)
	// Convert back to byte slice
	output := make([]byte, 4)
	binary.LittleEndian.PutUint32(output, indices)
	return output[:3]
}
