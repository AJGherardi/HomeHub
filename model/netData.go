package model

import (
	"reflect"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MakeNetData makes a new net data with the given webKey
func MakeNetData(webKey []byte) NetData {
	return NetData{
		NextGroupAddr:   0xc000,
		NextSceneNumber: 0x0001,
		WebKeys:         [][]byte{webKey},
	}
}

// NetData used for sending messages and adding new devices
type NetData struct {
	ID              primitive.ObjectID `bson:"_id"`
	NextGroupAddr   uint16
	NextSceneNumber uint16
	WebKeys         [][]byte
}

// GetNextGroupAddr returns the next group address
func (n *NetData) GetNextGroupAddr() uint16 {
	return n.NextGroupAddr
}

// GetNextSceneNumber returns the next scene number
func (n *NetData) GetNextSceneNumber() uint16 {
	return n.NextSceneNumber
}

// IncrementNextGroupAddr increments the next group address
func (n *NetData) IncrementNextGroupAddr() {
	n.NextGroupAddr++
}

// IncrementNextSceneNumber increments the next app key index
func (n *NetData) IncrementNextSceneNumber() {
	n.NextSceneNumber++
}

// CheckWebKey checks the validity of the given webKey
func (n *NetData) CheckWebKey(webKey []byte) bool {
	keys := n.WebKeys
	for _, key := range keys {
		if reflect.DeepEqual(key, webKey) {
			return true
		}
	}
	return false
}

// AddWebKey checks the validity of the given webKey
func (n *NetData) AddWebKey(webKey []byte) {
	n.WebKeys = append(n.WebKeys, webKey)
}
