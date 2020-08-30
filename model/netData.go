package model

import (
	"reflect"

	"github.com/AJGherardi/HomeHub/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func MakeNetData(webKey []byte, db DB) {
	netData := NetData{
		ID:              primitive.NewObjectID(),
		NextAppKeyIndex: []byte{0x01, 0x00},
		NextGroupAddr:   []byte{0xc0, 0x00},
		WebKeys:         [][]byte{webKey},
	}
	db.InsertNetData(netData)
}

// NetData used for sending messages and adding new devices
type NetData struct {
	ID              primitive.ObjectID `bson:"_id"`
	NextAppKeyIndex []byte
	NextGroupAddr   []byte
	WebKeys         [][]byte
}

func (n *NetData) GetNextAppKeyIndex() []byte {
	return n.NextAppKeyIndex
}

func (n *NetData) GetNextGroupAddr() []byte {
	return n.NextGroupAddr
}

func (n *NetData) IncrementNextGroupAddr(db DB) {
	n.NextGroupAddr = utils.IncrementAddr(n.NextGroupAddr)
	db.UpdateNetData(*n)
}

func (n *NetData) IncrementNextAppKeyIndex(db DB) {
	n.NextAppKeyIndex = utils.IncrementAddr(n.NextAppKeyIndex)
	db.UpdateNetData(*n)
}

func (n *NetData) CheckWebKey(webKey []byte) bool {
	keys := n.WebKeys
	for _, key := range keys {
		if reflect.DeepEqual(key, webKey) {
			return true
		}
	}
	return false
}
