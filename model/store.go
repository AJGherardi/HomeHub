package model

// Store is a capture of all data in use by the home hub
type Store struct {
	Groups  map[uint16]*Group
	NetData *NetData
}
