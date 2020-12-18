package model

import "errors"

// Store is a capture of all data in use by the home hub
type Store struct {
	Groups  map[uint16]*Group
	NetData *NetData
}

// SetNetData replaces the netdata
func (s *Store) SetNetData(netData NetData) {
	s.NetData = &netData
}

// GetNetData gets a ref to the net data
func (s *Store) GetNetData() (*NetData, error) {
	netData := s.NetData
	if netData == nil {
		return nil, errors.New("Not initialized")
	}
	return netData, nil
}

// GetGroup gets the ref to group with given addr
func (s *Store) GetGroup(groupAddr uint16) (*Group, error) {
	group := s.Groups[groupAddr]
	if group == nil {
		return nil, errors.New("No group with given address")
	}
	return group, nil
}

// AddGroup adds a group to the store
func (s *Store) AddGroup(addr uint16, group Group) {
	s.Groups[addr] = &group
}
