package stats

import (
	"encoding/json"
	"io/ioutil"
)

func StoreToFile(filename string) error {
	json, err := json.Marshal(runtimeStats)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, json, 0644)
	return err
}

func ReadFromFile(filename string) error {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw, &runtimeStats)
	return err
}
