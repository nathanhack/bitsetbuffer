package bitsetbuffer

import (
	"encoding/binary"
	"fmt"
	"io"
)

type BitSetReader interface {
	io.Reader
	ReadBits(bits []bool) (n int, err error)
}

type BitSetWriter interface {
	io.Writer
	WriteBits(bits []bool) (n int, err error)
}

type BitSetBuffer struct {
	pos int
	Set []bool
}

func NewFromBytes(bytes []byte) (*BitSetBuffer, error) {
	b := BitSetBuffer{}
	defer func() {
		b.pos = 0
	}()
	n, err := b.Write(bytes)
	if err != nil {
		return nil, err
	}
	if n != len(bytes) {
		return nil, fmt.Errorf("expected to store %v bytes in buffer but stored %v instead", len(bytes), n)
	}
	return &b, nil
}

func NewFromBits(bits []bool) (*BitSetBuffer, error) {
	b := BitSetBuffer{}
	defer func() {
		b.pos = 0
	}()
	n, err := b.WriteBits(bits)
	if err != nil {
		return nil, err
	}
	if n != len(bits) {
		return nil, fmt.Errorf("expected to store %v bits in buffer but stored %v instead", len(bits), n)
	}
	return &b, nil
}

func (bsb *BitSetBuffer) ResetToStart() {
	bsb.pos = 0
}

func (bsb *BitSetBuffer) ResetToEnd() {
	bsb.pos = len(bsb.Set)
}

func (bsb *BitSetBuffer) PosAtEnd() bool {
	return bsb.pos == len(bsb.Set)
}

func (bsb *BitSetBuffer) Read(bytes []byte) (n int, err error) {
	if bytes == nil {
		return 0, fmt.Errorf("error nil passed in")
	}
	if bsb.PosAtEnd() {
		return 0, io.EOF
	}

	n = 0
	for n < len(bytes) && !bsb.PosAtEnd() {
		bytes[n] = bsb.readByte()
		n++
	}
	return n, nil
}

func (bsb *BitSetBuffer) ReadBit() (bit bool, err error) {
	if len(bsb.Set) <= bsb.pos {
		return false, fmt.Errorf("no bits to read")
	}

	bit = bsb.Set[bsb.pos]
	bsb.pos++
	return bit, nil
}

func (bsb *BitSetBuffer) ReadBits(bits []bool) (n int, err error) {
	if bits == nil {
		return 0, fmt.Errorf("error nil passed in")
	}
	for n = 0; n < len(bits); n++ {
		if bsb.pos >= len(bsb.Set) {
			return
		}
		bits[n] = bsb.Set[bsb.pos]
		bsb.pos++
	}
	return
}

func (bsb *BitSetBuffer) Write(bytes []byte) (n int, err error) {
	n = 0
	for n < len(bytes) {
		bsb.writeByte(bytes[n])
		n++
	}
	return n, nil
}

func (bsb *BitSetBuffer) WriteBit(bit bool) {
	if bsb.Set == nil {
		bsb.Set = make([]bool, 0)
	}

	if bsb.pos < len(bsb.Set) {
		bsb.Set[bsb.pos] = bit
	} else {
		bsb.Set = append(bsb.Set, bit)
	}
	bsb.pos++
}

func (bsb *BitSetBuffer) WriteBits(bits []bool) (n int, err error) {
	if bsb.Set == nil {
		bsb.Set = make([]bool, 0)
	}

	for n = 0; n < len(bits); n++ {
		bsb.WriteBit(bits[n])
	}
	return
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (bsb *BitSetBuffer) readByte() (b byte) {
	b = 0
	defer func() {
		bsb.pos = min(len(bsb.Set), bsb.pos+8)
	}()

	for i := 0; i < 8; i++ {
		index := i + bsb.pos
		if index >= len(bsb.Set) {
			return
		}
		if bsb.Set[index] {
			b = b | (1 << i)
		}
	}
	return
}

func (bsb *BitSetBuffer) writeByte(b byte) {
	for i := 0; i < 8; i++ {
		value := b&(1<<i) > 0

		if bsb.pos < len(bsb.Set) {
			bsb.Set[bsb.pos] = value
		} else {
			if bsb.Set == nil {
				bsb.Set = make([]bool, 0)
			}
			bsb.Set = append(bsb.Set, value)
		}
		bsb.pos++
	}
}

func (bsb *BitSetBuffer) Bytes() []byte {
	old := bsb.pos
	defer func() {
		bsb.pos = old
	}()
	bsb.ResetToStart()
	buf := make([]byte, 0)
	for !bsb.PosAtEnd() {
		buf = append(buf, bsb.readByte())
	}
	return buf
}

func WriteUint(buf BitSetWriter, numOfBits int, endianness binary.ByteOrder, value uint64) error {
	intBits := make([]bool, numOfBits)
	for i := 0; i < numOfBits; i++ {
		intBits[i] = value&(1<<i) > 0
	}

	if endianness == binary.BigEndian {
		intBits = byteSwapBitsToBig(intBits)
	}

	n, err := buf.WriteBits(intBits)
	if err != nil {
		return err
	}
	if n != numOfBits {
		return fmt.Errorf("only %v of %v bits written", n, numOfBits)
	}
	return nil
}

func ReadUint(buf BitSetReader, numOfBits int, endianness binary.ByteOrder) (uint64, error) {
	intBits := make([]bool, numOfBits)
	n, err := buf.ReadBits(intBits)
	if err != nil {
		return 0, err
	}
	if n != numOfBits {
		return 0, fmt.Errorf("only %v of %v bits read", n, numOfBits)
	}

	if endianness == binary.BigEndian {
		intBits = byteSwapBitsFromBig(intBits)
	}

	value := uint64(0)
	for i := 0; i < numOfBits; i++ {
		if intBits[i] {
			value += 1 << i
		}
	}
	return value, nil
}

func WriteInt(buf BitSetWriter, numOfBits int, endianness binary.ByteOrder, value int64) error {
	intBits := make([]bool, numOfBits)
	for i := 0; i < numOfBits-1; i++ {
		intBits[i] = value&(1<<i) > 0
	}
	intBits[len(intBits)-1] = value < 0

	if endianness == binary.BigEndian {
		intBits = byteSwapBitsToBig(intBits)
	}

	n, err := buf.WriteBits(intBits)
	if err != nil {
		return err
	}
	if n != numOfBits {
		return fmt.Errorf("only %v of %v bits written", n, numOfBits)
	}
	return nil

}

func ReadInt(buf BitSetReader, numOfBits int, endianness binary.ByteOrder) (int64, error) {
	intBits := make([]bool, numOfBits)
	n, err := buf.ReadBits(intBits)
	if err != nil {
		return 0, err
	}
	if n != numOfBits {
		return 0, fmt.Errorf("only %v of %v bits read", n, numOfBits)
	}

	if endianness == binary.BigEndian {
		intBits = byteSwapBitsFromBig(intBits)
	}

	value := int64(0)
	if intBits[numOfBits-1] {
		value -= 1
	}
	for i := 0; i < numOfBits-1; i++ {
		value &= ^(1 << i)
		if intBits[i] {
			value |= 1 << i
		}
	}

	return value, nil
}

func byteSwapBitsToBig(bytes []bool) []bool {
	lenBytes := len(bytes)
	result := make([]bool, lenBytes)

	for start := 0; start < lenBytes; start = start + 8 {
		end := start + 8
		if end >= lenBytes {
			end = lenBytes
		}

		offset := lenBytes - end
		for j := 0; j < end-start; j++ {
			result[offset+j] = bytes[start+j]
		}
	}
	return result
}

func byteSwapBitsFromBig(bytes []bool) []bool {
	result := make([]bool, len(bytes))

	for i := 0; i < len(bytes); i++ {
		end := len(bytes) - i*8
		start := end - 8
		if start < 0 {
			start = 0
		}

		offset := i * 8
		for j := 0; j < end-start; j++ {
			result[offset+j] = bytes[start+j]
		}
	}
	return result
}
