package binaryutils

type BitStream struct {
	data   []byte
	offset int
	bit    uint8
}

func NewBitStream(data []byte) *BitStream {
	bs := new(BitStream)
	bs.data = data
	bs.bit = 0
	bs.offset = 0
	return bs
}

func (b BitStream) IsEnd() bool {
	return b.offset >= len(b.data)
}

func (b *BitStream) GetBit() uint8 {
	if b.bit > 7 {
		b.offset++
		if b.offset >= len(b.data) {
			return 0
		}
		b.bit = 0
	}
	abit := b.bit
	abyte := b.data[b.offset]
	value := (abyte & (1 << abit)) >> abit
	b.bit++
	return value
}

//TODO be more efficient if we have all the bits in the current loaded byte
func (b *BitStream) GetBits(length uint8) (value uint8) {
	var i, accum uint8
	if length > 8 {
		return 0
	}
	accum = 0
	for i = 0; i < length; i++ {
		bit := b.GetBit()
		accum += bit << i
	}
	return accum
}
