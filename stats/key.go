package stats

import "encoding/binary"

// Used as a key for maps
type coSchedCATKey struct {
	Application string
	CAT         uint64
}

func (k coSchedCATKey) MarshalText() ([]byte, error) {
	c := make([]byte, 8)
	binary.LittleEndian.PutUint64(c, k.CAT)

	a := []byte(k.Application)

	return append(c, a...), nil
}

func (k coSchedCATKey) UnmarshalText(text []byte) error {
	c := text[0:8]
	k.CAT = binary.LittleEndian.Uint64(c)

	k.Application = string(text[8:])

	return nil
}
