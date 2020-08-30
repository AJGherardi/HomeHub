package utils

import (
	"encoding/base64"
	"encoding/binary"
)

func EncodeBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

func DecodeBase64(input string) []byte {
	output, _ := base64.StdEncoding.DecodeString(input)
	return output
}

func IncrementAddr(input []byte) []byte {
	// Convert to uint16
	addr := binary.BigEndian.Uint16(input)
	// Increment
	addr++
	// Convert back to byte slice
	output := make([]byte, 2)
	binary.BigEndian.PutUint16(output, addr)
	return output
}

// IncrementKeyIndex increments an key index
func IncrementKeyIndex(input []byte) []byte {
	// Convert to uint16
	index := binary.LittleEndian.Uint16(input)
	// Increment
	index++
	// Convert back to byte slice
	output := make([]byte, 2)
	binary.LittleEndian.PutUint16(output, index)
	return output
}

func EncodeKeyIndices(netIndex []byte, appIndex []byte) []byte {
	// Remove Padding
	netInt := binary.LittleEndian.Uint16(netIndex)
	appInt := binary.LittleEndian.Uint16(appIndex)
	// Add indices
	indices := uint32(appInt)<<12 | uint32(netInt)
	// Convert back to byte slice
	output := make([]byte, 4)
	binary.LittleEndian.PutUint32(output, indices)
	return output[:3]
}

func RemoveDevAddr(slice [][]byte, i int) [][]byte {
	return append(slice[:i], slice[i+1:]...)
}
