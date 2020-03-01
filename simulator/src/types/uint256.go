package types

// Uint256 is 256 bit uint.
type Uint256 [32]byte

// Max returns max value of Uint256.
func (u Uint256) Max() Uint256 {
	var max [32]byte
	for i := 0; i < 32; i++ {
		max[i] = uint8(255)
	}
	return max
}

// AddUint8 adds u + a.
func (u Uint256) AddUint8(a int) Uint256 {
	newVal := [32]byte{}
	sum := uint16(u[0]) + uint16(uint8(a))
	newVal[0] = uint8(sum)
	carryUp := sum >> 8
	for i := 1; i < 32; i++ {
		sum = uint16(u[i]) + carryUp
		newVal[i] = uint8(sum)
		carryUp = sum >> 8
	}
	return newVal
}

// Add returns u + a.
func (u Uint256) Add(a Uint256) Uint256 {
	newVal := [32]byte{}
	carryUp := uint16(0)
	for i := 0; i < 32; i++ {
		sum := uint16(u[i]) + uint16(a[i]) + carryUp
		newVal[i] = uint8(sum)
		carryUp = sum >> 8
	}
	return newVal
}

// Divide2 perform as >> 1.
func (u Uint256) Divide2() Uint256 {
	newVal := [32]byte{}
	carryDown := uint8(0)
	for i := 0; i < 32; i++ {
		newVal[31-i] = (u[31-i] >> 1) + (carryDown << 7)
		carryDown = u[31-i] & uint8(1)
	}
	return newVal
}

// Larger returns u >= b.
func (u Uint256) Larger(b Uint256) bool {
	if u == b {
		return false
	}
	for i := 0; i < 32; i++ {
		if u[31-i] < b[31-i] {
			return false
		}
	}
	return true
}
