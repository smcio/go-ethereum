package bn256

import (
	"encoding/binary"
	"fmt"

	"github.com/holiman/uint256"
)

type gfP [4]uint64

func newGFp(x int64) (out *gfP) {
	if x >= 0 {
		out = &gfP{uint64(x)}
	} else {
		out = &gfP{uint64(-x)}
		gfpNeg(out, out)
	}

	montEncode(out, out)
	return out
}

func (e *gfP) String() string {
	return fmt.Sprintf("%16.16x%16.16x%16.16x%16.16x", e[3], e[2], e[1], e[0])
}

func (e *gfP) Set(f *gfP) {
	e[0] = f[0]
	e[1] = f[1]
	e[2] = f[2]
	e[3] = f[3]
}

func (e *gfP) Invert(f *gfP) {
	bits := [4]uint64{0x3c208c16d87cfd45, 0x97816a916871ca8d, 0xb85045b68181585d, 0x30644e72e131a029}

	sum, power := &gfP{}, &gfP{}
	sum.Set(rN1)
	power.Set(f)

	for word := 0; word < 4; word++ {
		for bit := uint(0); bit < 64; bit++ {
			if (bits[word]>>bit)&1 == 1 {
				gfpMul(sum, sum, power)
			}
			gfpMul(power, power, power)
		}
	}

	gfpMul(sum, sum, r3)
	e.Set(sum)
}

func (e *gfP) Marshal(out []byte) {
	for w := uint(0); w < 4; w++ {
		for b := uint(0); b < 8; b++ {
			out[8*w+b] = byte(e[3-w] >> (56 - 8*b))
		}
	}
}

var (
	flagInf  = (1 << 6)
	flagNeg  = (1 << 7)
	flagMask = 0x3F
	p        = uint256.Int{0x3c208c16d87cfd47, 0x97816a916871ca8d, 0xb85045b68181585d, 0x30644e72e131a029}
)

func (e *gfP) Unmarshal(in []byte, isInf, isNeg *bool) error {
	var maskedInput [32]byte
	copy(maskedInput[:], in)

	if isInf != nil {
		infSet := maskedInput[0]&byte(flagInf) != 0
		negSet := maskedInput[0]&byte(flagNeg) != 0
		if infSet && negSet {
			return fmt.Errorf("both inf and neg flags set")
		}
		*isInf = infSet
		if isNeg != nil {
			*isNeg = negSet
		}
		maskedInput[0] &= byte(flagMask)
	}

	var v uint256.Int
	v[3] = binary.BigEndian.Uint64(maskedInput[0:8])
	v[2] = binary.BigEndian.Uint64(maskedInput[8:16])
	v[1] = binary.BigEndian.Uint64(maskedInput[16:24])
	v[0] = binary.BigEndian.Uint64(maskedInput[24:32])
	if v.Cmp(&p) >= 0 {
		return fmt.Errorf("value is larger than or equal to the field modulus")
	}

	var newGfp gfP
	newGfp[3] = binary.BigEndian.Uint64(maskedInput[0:8])
	newGfp[2] = binary.BigEndian.Uint64(maskedInput[8:16])
	newGfp[1] = binary.BigEndian.Uint64(maskedInput[16:24])
	newGfp[0] = binary.BigEndian.Uint64(maskedInput[24:32])

	*e = newGfp
	return nil
}

func montEncode(c, a *gfP) { gfpMul(c, a, r2) }
func montDecode(c, a *gfP) { gfpMul(c, a, &gfP{1}) }
