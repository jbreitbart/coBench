package bit

// Sets the bit at pos in the integer n.
func Set(val uint64, pos uint64) uint64 {
	// TODO error if pos >64
	val |= (1 << pos)
	return val
}

// Clears the bit at pos in n.
func Clear(val uint64, pos uint64) uint64 {
	// TODO error if pos >64
	var mask uint64
	mask = ^(1 << pos)
	val &= mask
	return val
}

// Has checks if a bit is set
func Has(val uint64, pos uint64) bool {
	// TODO error if pos >64
	v := val & (1 << pos)
	return (v > 0)
}

func SetFirstN(val uint64, n uint64) uint64 {
	// TODO error if n >64; n<0
	for i := (uint64)(0); i < n; i++ {
		val = Set(val, i)
	}
	return val
}

func SetLastN(val uint64, n uint64, size uint64) uint64 {
	// TODO error if n >64, size >64; <0
	for i := (uint64)(size - 1); i >= n; i-- {
		val = Set(val, i)
	}
	return val

}
