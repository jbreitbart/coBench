package stats

import (
	"encoding/binary"
	"encoding/json"
	"io/ioutil"
)

func (k coSchedCATKey) MarshalText() ([]byte, error) {
	c := make([]byte, 8)
	binary.LittleEndian.PutUint64(c, k.CAT)

	a := []byte(k.Application)

	return append(c, a...), nil
}

func (k *coSchedCATKey) UnmarshalText(text []byte) error {
	c := text[:8]
	k.CAT = binary.LittleEndian.Uint64(c)

	k.Application = string(text[8:])
	// TODO understand why there are 0bytes! Probably a bug...
	for i := 8; i < len(text); i++ {
		if text[i] != 0 {
			k.Application = string(text[i:])
			break
		}
	}

	return nil
}

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
