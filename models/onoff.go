package models

// OnOffGet makes an generic onoff get payload
func OnOffGet() []byte {
	opcode := []byte{0x82, 0x01}
	return opcode
}

// OnOffSet makes an generic onoff set payload
func OnOffSet(onoff bool) []byte {
	opcode := []byte{0x82, 0x02}
	var value byte
	if onoff == true {
		value = 0x01
	} else {
		value = 0x00
	}
	output := append(opcode, []byte{value, 0x00}...)
	return output
}
