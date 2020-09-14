package utils

import (
	"encoding/base64"
	"encoding/binary"
	"reflect"
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

// Increment16 increments the given two byte array
func Increment16(input []byte) []byte {
	// Convert to uint16
	addr := binary.LittleEndian.Uint16(input)
	// Increment
	addr++
	// Convert back to byte slice
	output := make([]byte, 2)
	binary.LittleEndian.PutUint16(output, addr)
	return output
}

// RemoveDevAddr removes the address at the given index
func RemoveDevAddr(slice [][]byte, i int) [][]byte {
	return append(slice[:i], slice[i+1:]...)
}

// Delete removes the element at the given index
func Delete(arr interface{}, index int) {
	vField := reflect.ValueOf(arr)
	value := vField.Elem()
	if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
		result := reflect.AppendSlice(value.Slice(0, index), value.Slice(index+1, value.Len()))
		value.Set(result)
	}
}
