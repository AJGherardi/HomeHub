package main

import "go.mongodb.org/mongo-driver/bson/primitive"

// Group holds a collection of devices and its app key id
type Group struct {
	Name     string
	Aid      []byte
	Addr     []byte
	DevAddrs [][]byte
}

// Device holds the name addr and type of device
type Device struct {
	Name     string
	Type     string
	Addr     []byte
	Seq      []byte
	Elements []Element
}

// Element holds an elements addr and its state
type Element struct {
	Addr  []byte
	State State
}

// State has a value and a type
type State struct {
	State     []byte
	StateType string
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
