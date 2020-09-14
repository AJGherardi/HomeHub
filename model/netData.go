package model

import (
	"reflect"

	"github.com/AJGherardi/HomeHub/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MakeNetData makes a new net data with the given webKey
func MakeNetData(webKey []byte, db DB) {
	netData := NetData{
		ID:              primitive.NewObjectID(),
		NextGroupAddr:   []byte{0x00, 0xc0},
		NextSceneNumber: []byte{0x01, 0x00},
		WebKeys:         [][]byte{webKey},
	}
	db.InsertNetData(netData)
}

// NetData used for sending messages and adding new devices
type NetData struct {
	ID              primitive.ObjectID `bson:"_id"`
	NextGroupAddr   []byte
	NextSceneNumber []byte
	WebKeys         [][]byte
}

// GetNextGroupAddr returns the next group address
func (n *NetData) GetNextGroupAddr() []byte {
	return n.NextGroupAddr
}

// GetNextSceneNumber returns the next scene number
func (n *NetData) GetNextSceneNumber() []byte {
	return n.NextSceneNumber
}

// IncrementNextGroupAddr incrments the next group address
func (n *NetData) IncrementNextGroupAddr(db DB) {
	n.NextGroupAddr = utils.Increment16(n.NextGroupAddr)
	db.UpdateNetData(*n)
}

// IncrementNextSceneNumber incrments the next app key index
func (n *NetData) IncrementNextSceneNumber(db DB) {
	n.NextSceneNumber = utils.Increment16(n.NextSceneNumber)
	db.UpdateNetData(*n)
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
