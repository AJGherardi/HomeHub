package main

import (
	"encoding/base64"
	"encoding/binary"
)

func encodeBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

func decodeBase64(imput string) []byte {
	output, _ := base64.StdEncoding.DecodeString(imput)
	return output
}

func incrementAddr(imput []byte) []byte {
	// Convert to uint16
	addr := binary.BigEndian.Uint16(imput)
	// Increment
	addr++
	// Convert back to byte slice
	output := make([]byte, 2)
	binary.BigEndian.PutUint16(output, addr)
	return output
}

// IncrementKeyIndex increments an key index
func incrementKeyIndex(imput []byte) []byte {
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

func removeDevAddr(slice [][]byte, i int) [][]byte {
	return append(slice[:i], slice[i+1:]...)
}
