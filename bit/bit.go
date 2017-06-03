package bit

// Sets the bit at pos in the integer n.
func Set(val int64, pos uint64) int64 {
	val |= (1 << pos)
	return val
}

// Clears the bit at pos in n.
func Clear(val int64, pos uint64) int64 {
	var mask int64
	mask = ^(1 << pos)
	val &= mask
	return val
}

// Has checks if a bit is set
func Has(val int64, pos uint64) bool {
	v := val & (1 << pos)
	return (v > 0)
}

func SetFirstN(val int64, n int) int64 {
	// TODO error if n >64; n<0
	for i := 0; i < n; i++ {
		val = Set(val, (uint64)(i))
	}
	return val
}

func SetLastN(val int64, n int, size int) int64 {
	// TODO error if n >64, size >64; <0
	for i := size; i >= n; i-- {
		val = Set(val, (uint64)(i))
	}
	return val

}
