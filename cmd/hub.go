package cmd

import (
	"crypto/rand"
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	mesh "github.com/AJGherardi/GoMeshController"
	"github.com/AJGherardi/HomeHub/model"
)

// ConfigHub Initlizes the hub for the first time
func ConfigHub(store *model.Store, controller mesh.Controller) []byte {
	// Make a web key
	webKey := make([]byte, 16)
	rand.Read(webKey)
	// Start data save threed
	go SaveStore(store)
	// Add and get net data
	netData := model.MakeNetData(webKey)
	store.NetData = &netData
	// Setup controller
	controller.Setup()
	time.Sleep(100 * time.Millisecond)
	// Add an app key
	controller.AddKey(0x0000)
	return webKey
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
func SaveToFile(store model.Store) {
	jsonData, _ := json.Marshal(store)
	ioutil.WriteFile("home.data", jsonData, 0777)
}

// ReadFromFile reads data from the home.data file
func ReadFromFile() model.Store {
	jsonData, _ := ioutil.ReadFile("home.data")
	store := new(model.Store)
	json.Unmarshal(jsonData, store)
	return *store
}

// SaveStore handles updating the store file
func SaveStore(store *model.Store) {
	for {
		os.Remove("home.data")
		jsonData, _ := json.Marshal(store)
		ioutil.WriteFile("home.data", jsonData, 0777)
		time.Sleep(500 * time.Millisecond)
	}
}

// AddAccessKey generates and saves a Web Key
func AddAccessKey(store *model.Store) []byte {
	// Make a web key
	webKey := make([]byte, 16)
	rand.Read(webKey)
	// Add webKey to netData
	netData := store.NetData
	netData.AddWebKey(webKey)
	return webKey
}
