package main

import (
	"reflect"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func makeNetData(netKey, webKey []byte) {
	netData := NetData{
		ID:              primitive.NewObjectID(),
		NetKey:          netKey,
		NetKeyIndex:     []byte{0x00, 0x00},
		NextAppKeyIndex: []byte{0x01, 0x00},
		Flags:           []byte{0x00},
		IvIndex:         []byte{0x00, 0x00, 0x00, 0x01},
		NextAddr:        []byte{0x00, 0x01},
		NextGroupAddr:   []byte{0xc0, 0x00},
		HubSeq:          []byte{0x00, 0x00, 0x00},
		WebKeys:         [][]byte{webKey},
	}
	insertNetData(netCollection, netData)
}

// NetData used for sending msgs and adding new devices
type NetData struct {
	ID              primitive.ObjectID `bson:"_id"`
	NetKey          []byte
	NetKeyIndex     []byte
	NextAppKeyIndex []byte
	Flags           []byte
	IvIndex         []byte
	NextAddr        []byte
	NextGroupAddr   []byte
	HubSeq          []byte
	WebKeys         [][]byte
}

func (n *NetData) getNetKey() []byte {
	return n.NetKey
}

func (n *NetData) getNetKeyIndex() []byte {
	return n.NetKeyIndex
}

func (n *NetData) getNextAppKeyIndex() []byte {
	return n.NextAppKeyIndex
}

func (n *NetData) getFlags() []byte {
	return n.Flags
}

func (n *NetData) getIvIndex() []byte {
	return n.IvIndex
}

func (n *NetData) getNextAddr() []byte {
	return n.NextAddr
}

func (n *NetData) getHubSeq() []byte {
	return n.HubSeq
}

func (n *NetData) getNextGroupAddr() []byte {
	return n.NextGroupAddr
}

func (n *NetData) updateNextAddr(addr []byte) {
	n.NextAddr = addr
	updateNetData(netCollection, *n)
}

func (n *NetData) updateHubSeq(seq []byte) {
	n.HubSeq = seq
	updateNetData(netCollection, *n)
}

func (n *NetData) incrementNextGroupAddr() {
	n.NextGroupAddr = incrementAddr(n.NextGroupAddr)
	updateNetData(netCollection, *n)
}

func (n *NetData) incrementNextAppKeyIndex() {
	n.NextAppKeyIndex = incrementAddr(n.NextAppKeyIndex)
	updateNetData(netCollection, *n)
}

func (n *NetData) checkWebKey(webKey []byte) bool {
	keys := n.WebKeys
	for _, key := range keys {
		if reflect.DeepEqual(key, webKey) {
			return true
		}
	}
	return false
}
