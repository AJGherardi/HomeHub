package main

// NodeReset resets a node
func nodeReset() []byte {
	opcode := []byte{0x80, 0x49}
	return opcode
}

// AppKeyAdd makes an appkey add payload
func appKeyAdd(netIndex []byte, appIndex []byte, appKey []byte) []byte {
	opcode := []byte{0x00}
	indices := encodeKeyIndices(netIndex, appIndex)
	payload := append(opcode, indices...)
	payload = append(payload, appKey...)
	return payload
}

// AppKeyBind makes an appkey bind payload
func appKeyBind(addr []byte, appIndex []byte, modelID []byte) []byte {
	opcode := []byte{0x80, 0x3d}
	elemAddr := []byte{addr[1], addr[0]}
	model := []byte{modelID[1], modelID[0]}
	payload := append(opcode, elemAddr...)
	payload = append(payload, appIndex...)
	payload = append(payload, model...)
	return payload
}

// ConfigDataGet makes an config data get payload
func configDataGet() []byte {
	opcode := []byte{0x80, 0x50}
	return opcode
}
