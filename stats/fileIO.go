package stats

import (
	"io/ioutil"
)

// StoreToFile stores the current stats as json in a file
func StoreToFile(filename string) error {
	json, err := CreateJSON()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, json, 0644)
	return err
}

// ReadFromFile reads a json file stored by StoreToFile and updates the local state of the package
func ReadFromFile(filename string) error {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = StoreJSON(raw)
	return err
}
