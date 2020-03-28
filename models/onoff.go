package models

// OnOffGet makes an generic onoff get payload
func OnOffGet() []byte {
	opcode := []byte{0x82, 0x01}
	return opcode
}

// OnOffSet makes an generic onoff set payload
func OnOffSet(onoff byte) []byte {
	opcode := []byte{0x82, 0x02}

	output := append(opcode, []byte{onoff, 0x00}...)
	return output
}
