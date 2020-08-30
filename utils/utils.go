package utils

import (
	"encoding/base64"
	"encoding/binary"
)

// EncodeBase64 converts bytes to base 64
func EncodeBase64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

// DecodeBase64 converts base 64 to bytes
func DecodeBase64(input string) []byte {
	output, _ := base64.StdEncoding.DecodeString(input)
	return output
}

// IncrementAddr increments the given addr
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

// IncrementKeyIndex increments the given key index
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

// RemoveDevAddr removes the address at the given index
func RemoveDevAddr(slice [][]byte, i int) [][]byte {
	return append(slice[:i], slice[i+1:]...)
}
