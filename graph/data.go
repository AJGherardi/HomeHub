package graph

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/AJGherardi/HomeHub/model"
)

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
