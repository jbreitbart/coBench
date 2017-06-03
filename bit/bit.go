package bit

// Sets the bit at pos in the integer n.
func Set(n int64, pos uint64) int64 {
	n |= (1 << pos)
	return n
}

// Clears the bit at pos in n.
func Clear(n int64, pos uint64) int64 {
	var mask int64
	mask = ^(1 << pos)
	n &= mask
	return n
}

func Has(n int64, pos uint64) bool {
	val := n & (1 << pos)
	return (val > 0)
}
