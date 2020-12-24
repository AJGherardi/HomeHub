package model

import (
	"errors"
	"reflect"
)

// Store is a capture of all data in use by the home hub
type Store struct {
	Groups          map[uint16]*Group
	NextGroupAddr   uint16
	NextSceneNumber uint16
	WebKeys         [][]byte
	Configured      bool
}

// MakeStore makes a new store with the given webKey
func MakeStore() Store {
	return Store{
		NextGroupAddr:   0xc000,
		NextSceneNumber: 0x0001,
		WebKeys:         [][]byte{},
		Groups:          map[uint16]*Group{},
		Configured:      false,
	}
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

// GetNextGroupAddr returns the next group address
func (s *Store) GetNextGroupAddr() uint16 {
	return s.NextGroupAddr
}

// GetNextSceneNumber returns the next scene number
func (s *Store) GetNextSceneNumber() uint16 {
	return s.NextSceneNumber
}

// IncrementNextGroupAddr increments the next group address
func (s *Store) IncrementNextGroupAddr() {
	s.NextGroupAddr++
}

// IncrementNextSceneNumber increments the next app key index
func (s *Store) IncrementNextSceneNumber() {
	s.NextSceneNumber++
}

// GetConfigured returns a bool indicating if the network is configured
func (s *Store) GetConfigured() bool {
	return s.Configured
}

// MarkConfigured marks the network as configured
func (s *Store) MarkConfigured() {
	s.Configured = true
}

// CheckWebKey checks the validity of the given webKey
func (s *Store) CheckWebKey(webKey []byte) bool {
	keys := s.WebKeys
	for _, key := range keys {
		if reflect.DeepEqual(key, webKey) {
			return true
		}
	}
	return false
}

// AddWebKey checks the validity of the given webKey
func (s *Store) AddWebKey(webKey []byte) {
	s.WebKeys = append(s.WebKeys, webKey)
}
