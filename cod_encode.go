package ecs

import (
	"github.com/unitoftime/cod/backend"
)

func (t Id) CodEquals(tt Id) bool {
	return t == tt
}

func (t Id) EncodeCod(bs []byte) []byte {

	{
		value0 := uint32(t)

		bs = backend.WriteVarUint32(bs, value0)

	}
	return bs
}

func (t *Id) DecodeCod(bs []byte) (int, error) {
	var err error
	var n int
	var nOff int

	{
		var value0 uint32

		value0, nOff, err = backend.ReadVarUint32(bs[n:])
		if err != nil {
			return 0, err
		}
		n += nOff

		*t = Id(value0)
	}

	return n, err
}
