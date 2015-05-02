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

func (b *BitStream) GetBits(length uint8) (value uint8) {
	var i uint8
	if length > 8 {
		return 0
	}
	if length+b.bit < 8 {
		abyte := b.data[b.offset]
		value = (abyte >> b.bit) & (1<<length - 1)
		b.bit += length
	} else {
		value = 0
		for i = 0; i < length; i++ {
			bit := b.GetBit()
			value += bit << i
		}
	}
	return
}

func ByteToBits(in byte) (out [8]uint8) {
	for x := 0; x < 8; x++ {
		if in&(0x80>>uint8(x)) != 0 {
			out[x] = 1
		} else {
			out[x] = 0
		}
	}
	return
}
