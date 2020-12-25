package cmd

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"time"

	mesh "github.com/AJGherardi/GoMeshController"
	"github.com/AJGherardi/HomeHub/model"
)

// Network holds all needed data for the operation of the mesh network
type Network struct {
	Store              *model.Store
	Controller         mesh.Controller
	UnprovisionedNodes *[][]byte
	NodeAdded          chan uint16
}

// ReadFunc is a fuction that can be called on a seprate thred that will read from the mesh controller
type ReadFunc func(updateState func(), publishEvents func(addr uint16))

// WriteFunc is a fuction that writes the store to a file every 500 ms
type WriteFunc func()

// MakeNetwork opens all resources neaded to run the network and returns read and write functions that need to be run on their own goroutines
func MakeNetwork() (Network, ReadFunc, WriteFunc) {
	// Try to read the store from a file
	store, err := ReadFromFile()
	// If the store is unreadable make a new store
	if err != nil {
		store = model.MakeStore()
	}
	// Open Mesh Controller
	controller, err := mesh.Open()
	if err != nil {
		panic("Can not open mesh controller")
	}
	// Make network object
	network := Network{
		Store:              &store,
		Controller:         controller,
		UnprovisionedNodes: &[][]byte{},
		NodeAdded:          make(chan uint16),
	}
	// Return read and write functions
	return network,
		func(updateState func(), publishEvents func(addr uint16)) {
			err := network.Controller.Read(
				// onSetupStatus
				func() {},
				// onAddKeyStatus
				func(appIdx uint16) {},
				// onUnprovisionedBeacon
				func(uuid []byte) {
					*network.UnprovisionedNodes = append(*network.UnprovisionedNodes, uuid)
				},
				// onNodeAdded
				func(addr uint16) {
					network.NodeAdded <- addr
				},
				// onState
				func(addr uint16, state byte) {
					network.UpdateState(addr, state)
					// Push new state
					updateState()
				},
				// onEvent
				func(addr uint16) {
					publishEvents(addr)
				},
			)
			if err != nil {
				panic("Can not read from mesh controller")
			}
		},
		func() { SaveStore(&store) }
}

// ConfigHub Initlizes the hub for the first time
func (n *Network) ConfigHub() ([]byte, error) {
	// Make a web key
	webKey := make([]byte, 16)
	rand.Read(webKey)
	// Start data save threed
	go SaveStore(n.Store)
	// Add webKey and mark as configured
	n.Store.AddWebKey(webKey)
	n.Store.MarkConfigured()
	// Setup controller
	sendErr := n.Controller.Setup()
	if sendErr != nil {
		return []byte{}, errors.New("Can not setup controller")
	}
	time.Sleep(100 * time.Millisecond)
	// Add an app key
	sendErr = n.Controller.AddKey(0x0000)
	if sendErr != nil {
		return []byte{}, errors.New("Can not setup controller")
	}
	return webKey, nil
}

// ResetHub Initlizes the hub for the first time
func (n *Network) ResetHub() {
	// Remove all devices
	groups := n.Store.Groups
	for i := range groups {
		n.RemoveGroup(i)
	}
	// TODO: Clean house
	// Reset mesh controller
	n.Controller.Reset()
	time.Sleep(time.Second)
	n.Controller.Reboot()
	go func() {
		time.Sleep(300 * time.Millisecond)
		os.Exit(0)
	}()
}

// CheckIfConfigured checks if the network is configured
func (n *Network) CheckIfConfigured() bool {
	return n.Store.GetConfigured()
}

// AddAccessKey generates and saves a Web Key
func (n *Network) AddAccessKey() ([]byte, error) {
	// Make a web key
	webKey := make([]byte, 16)
	rand.Read(webKey)
	// Add webKey to store
	n.Store.AddWebKey(webKey)
	return webKey, nil
}

// CheckAccessKey checks the access key using know access keys
func (n *Network) CheckAccessKey(webKey []byte) bool {
	return n.Store.CheckWebKey(webKey)
}

// GetUnprovisionedNodes returns a list of detected unprovisioned nodes
func (n *Network) GetUnprovisionedNodes() [][]byte {
	return *n.UnprovisionedNodes
}

// Close closes all of the resorces that the network depends on
func (n *Network) Close() {
	n.Controller.Close()
}

// SaveToFile writes data to the home.data file
func SaveToFile(store model.Store) error {
	jsonData, marshalErr := json.Marshal(store)
	if marshalErr != nil {
		return errors.New("Failed to marshal store")
	}
	writeErr := ioutil.WriteFile("home.data", jsonData, 0777)
	if writeErr != nil {
		return errors.New("Failed to write to file")
	}
	return nil
}

// ReadFromFile reads data from the home.data file
func ReadFromFile() (model.Store, error) {
	jsonData, readErr := ioutil.ReadFile("home.data")
	if readErr != nil {
		return model.Store{}, errors.New("Failed to read from file")
	}
	store := new(model.Store)
	unmarshalErr := json.Unmarshal(jsonData, store)
	if unmarshalErr != nil {
		return model.Store{}, errors.New("Failed to unmarshal store")
	}
	return *store, nil
}

// SaveStore handles updating the store file
func SaveStore(store *model.Store) {
	for {
		jsonData, marshalErr := json.Marshal(store)
		if marshalErr != nil {
			continue
		}
		os.Remove("home.data")
		ioutil.WriteFile("home.data", jsonData, 0777)
		time.Sleep(500 * time.Millisecond)
	}
}
