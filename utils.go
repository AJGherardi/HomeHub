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
