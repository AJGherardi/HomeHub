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

func encodeProvData(netKey []byte, keyIndex []byte, flags []byte, ivIndex []byte, devAddr []byte) ProvData {
	keyB64 := encodeBase64(netKey)
	indexB64 := encodeBase64(keyIndex)
	flagsB64 := encodeBase64(flags)
	ivIndexB64 := encodeBase64(ivIndex)
	devAddrB64 := encodeBase64(devAddr)
	return ProvData{
		NetworkKey:  keyB64,
		KeyIndex:    indexB64,
		Flags:       flagsB64,
		IvIndex:     ivIndexB64,
		NextDevAddr: devAddrB64,
	}
}

func incrementAddr(imput []byte) []byte {
	// Convert to uint16
	short := binary.BigEndian.Uint16(imput)
	// Increment
	short++
	// Convert back to byte slice
	output := make([]byte, 2)
	binary.BigEndian.PutUint16(output, short)
	return output
}
