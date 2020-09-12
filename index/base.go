package index

type Base uint

const (
	Base2 Base = 1 + iota
	Base4
	Base8
	Base16
	Base32
)

func (b Base) LeadingZeros(obj []byte) uint {
	if b == Base(0) {
		panic("Invalid base Base1")
	}
	numZeros := uint(0)
	ctr := Base(0)
	for _, bte := range obj {
		for i := 0; i < 8; i++ {
			if bte&byte(1<<(7-i)) != 0 {
				return numZeros
			}
			ctr++
			if ctr == b {
				numZeros++
				ctr = 0
			}
		}
	}
	return numZeros
}
