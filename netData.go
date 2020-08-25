package main

import (
	"reflect"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func makeNetData(webKey []byte) {
	netData := NetData{
		ID:              primitive.NewObjectID(),
		NextAppKeyIndex: []byte{0x01, 0x00},
		NextGroupAddr:   []byte{0xc0, 0x00},
		WebKeys:         [][]byte{webKey},
	}
	insertNetData(netCollection, netData)
}

// NetData used for sending msgs and adding new devices
type NetData struct {
	ID              primitive.ObjectID `bson:"_id"`
	NextAppKeyIndex []byte
	NextGroupAddr   []byte
	WebKeys         [][]byte
}

func (n *NetData) getNextAppKeyIndex() []byte {
	return n.NextAppKeyIndex
}

func (n *NetData) getNextGroupAddr() []byte {
	return n.NextGroupAddr
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
