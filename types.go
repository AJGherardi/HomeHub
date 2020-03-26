package main

// Device holds the name addr and type of device
type Device struct {
	Name string
	Type string
	Addr []byte
}

// NetData used for sending msgs and adding new devices
type NetData struct {
	NetKey      []byte
	NetKeyIndex []byte
	Flags       []byte
	IvIndex     []byte
	NextDevAddr []byte
	HubSeq      []byte
}

// ProvData holds all data needed to setup a device in base64
type ProvData struct {
	NetworkKey  string
	KeyIndex    string
	Flags       string
	IvIndex     string
	NextDevAddr string
}
