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

// ConfigHub Initlizes the hub for the first time
func ConfigHub(store *model.Store, controller mesh.Controller) ([]byte, error) {
	// Make a web key
	webKey := make([]byte, 16)
	rand.Read(webKey)
	// Start data save threed
	go SaveStore(store)
	// Add and get net data
	netData := model.MakeNetData(webKey)
	store.SetNetData(netData)
	// Setup controller
	sendErr := controller.Setup()
	if sendErr != nil {
		return []byte{}, errors.New("Can not setup controller")
	}
	time.Sleep(100 * time.Millisecond)
	// Add an app key
	sendErr = controller.AddKey(0x0000)
	if sendErr != nil {
		return []byte{}, errors.New("Can not setup controller")
	}
	return webKey, nil
}

// ResetHub Initlizes the hub for the first time
func ResetHub(store *model.Store, controller mesh.Controller) {
	// Remove all devices
	groups := store.Groups
	for i := range groups {
		RemoveGroup(store, controller, i)
	}
	// TODO: Clean house
	// Reset mesh controller
	controller.Reset()
	time.Sleep(time.Second)
	controller.Reboot()
	go func() {
		time.Sleep(300 * time.Millisecond)
		os.Exit(0)
	}()
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

// AddAccessKey generates and saves a Web Key
func AddAccessKey(store *model.Store) ([]byte, error) {
	// Make a web key
	webKey := make([]byte, 16)
	rand.Read(webKey)
	// Add webKey to netData
	netData, getErr := store.GetNetData()
	if getErr != nil {
		return []byte{}, getErr
	}
	netData.AddWebKey(webKey)
	return webKey, nil
}
