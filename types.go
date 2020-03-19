package main

// KeyType is either app or dev
type KeyType bool

// AppKey is a app key and its public aid
type AppKey struct {
	// For aid or dev address
	ID  []byte
	Key []byte
	// Key index is only for App Keys
	KeyIndex []byte
	Type     KeyType
}

// Device holds the name addr and type of device
type Device struct {
	Name string
	Addr string
	Type string
}

// NetData used for sending msgs and adding new devices
type NetData struct {
	NetKey      []byte
	NetKeyIndex []byte
	Flags       []byte
	IvIndex     []byte
	NextDevAddr []byte
}

// ProvData holds all data needed to setup a device in base64
type ProvData struct {
	NetworkKey  string
	KeyIndex    string
	Flags       string
	IvIndex     string
	NextDevAddr string
}

const (
	// KeyDev is a key used for dev msgs
	KeyDev KeyType = false
	// KeyApp is a key used for app msgs
	KeyApp KeyType = true
)
