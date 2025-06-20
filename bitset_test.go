package bitsetbuffer

import (
	"encoding/binary"
	"math/bits"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
)

func TestNewFromBytes(t *testing.T) {
	b, err := NewFromBytes([]byte{0, 1})
	if err != nil {
		t.Fatal(err)
	}
	expected := []bool{false, false, false, false, false, false, false, false, true, false, false, false, false, false, false, false}
	actual := b.Set
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected \n%v\n but foudn \n%v\n", expected, actual)
	}
}

func TestNewFromBits(t *testing.T) {
	b, err := NewFromBits([]bool{true, false, true})
	if err != nil {
		t.Fatal(err)
	}
	expected := []bool{true, false, true}
	actual := b.Set
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected \n%v\n but foudn \n%v\n", expected, actual)
	}
}

func TestBitSetBuffer_Read(t *testing.T) {
	b, err := NewFromBytes([]byte{0, 1})
	if err != nil {
		t.Fatal(err)
	}
	expected := []byte{0, 1}

	actual := make([]byte, 2)
	n, err := b.Read(actual)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("expected 2 but found %v", n)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected \n%v\n but foudn \n%v\n", expected, actual)
	}
}

func TestBitSetBuffer_WriteBits(t *testing.T) {
	b := BitSetBuffer{}
	expectedBits := []bool{true, false, true}
	n, err := b.WriteBits(expectedBits)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(expectedBits) {
		t.Fatalf("expected %v but found %v", len(expectedBits), n)
	}

	actual := b.Set
	if !reflect.DeepEqual(actual, expectedBits) {
		t.Fatalf("expected \n%v\n but foudn \n%v\n", expectedBits, actual)
	}

	b.ResetToStart()
	bytes := make([]byte, 2) //1 more than we should get
	n, err = b.Read(bytes)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected %v but found %v", 1, n)
	}
	bytes = bytes[:1]
	expectedBytes := []byte{5}
	if !reflect.DeepEqual(expectedBytes, bytes) {
		t.Fatalf("expected \n%v\n but foudn \n%v\n", expectedBytes, bytes)
	}
}

func TestBitSetBuffer_ReadBits(t *testing.T) {
	expected := []bool{true, false, true}
	b, err := NewFromBits(expected)
	if err != nil {
		t.Fatal(err)
	}
	actual := make([]bool, 4)
	n, err := b.ReadBits(actual)
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("expected 3 but found %v", n)
	}
	actual = actual[:3]
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("expected \n%v\n but foudn \n%v\n", expected, actual)
	}
}

func TestBitSetBuffer_Read2(t *testing.T) {
	b := BitSetBuffer{}
	n, err := b.Read(nil)
	if err == nil {
		t.Fatalf("expected and error but found none")
	}
	if n != 0 {
		t.Fatalf("expected to find 0 but found %v", n)
	}
}

func TestBitSetBuffer_ReadBits2(t *testing.T) {
	b := BitSetBuffer{}
	n, err := b.ReadBits(nil)
	if err == nil {
		t.Fatalf("expected and error but found none")
	}
	if n != 0 {
		t.Fatalf("expected to find 0 but found %v", n)
	}
}

func TestBitSetBuffer_Write(t *testing.T) {
	b, err := NewFromBits([]bool{false, false, false, false})
	if err != nil {
		t.Fatal(err)
	}

	b.ResetToEnd()

	n, err := b.Write([]byte{0xff})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected 1 but found %v", n)
	}

	expectedBits := []bool{false, false, false, false, true, true, true, true, true, true, true, true}
	if !reflect.DeepEqual(b.Set, expectedBits) {
		t.Fatalf("expected \n%v\n but foudn \n%v\n", expectedBits, b.Set)
	}
	expectedBytes := []byte{0xf0, 0x0f}
	if !reflect.DeepEqual(b.Bytes(), expectedBytes) {
		t.Fatalf("expected \n%v\n but foudn \n%v\n", expectedBytes, b.Bytes())
	}
}

func TestByteSwapBits(t *testing.T) {
	tests := []struct {
		expected    []bool
		expectedBig []bool
	}{
		{
			[]bool{
				true, true, false, false, true, false, true, true,
				false, false, true,
			},
			[]bool{
				false, false, true,
				true, true, false, false, true, false, true, true,
			},
		},

		{
			[]bool{
				true, true, false, false, true, false, true, true,
				false, false, true, true, true, false, false, false,
			},
			[]bool{
				false, false, true, true, true, false, false, false,
				true, true, false, false, true, false, true, true,
			},
		},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actualBig := byteSwapBitsToBig(test.expected)
			if !reflect.DeepEqual(test.expectedBig, actualBig) {
				t.Fatalf("expected \n%v\n but found \n%v\n", test.expectedBig, actualBig)
			}

			actual := byteSwapBitsFromBig(actualBig)

			if !reflect.DeepEqual(test.expected, actual) {
				t.Fatalf("expected \n%v\n but found \n%v\n", test.expected, actual)
			}
		})
	}
}

func TestWriteUint(t *testing.T) {
	type Input struct {
		numOfBits int
		values    []uint64
	}
	tests := []struct {
		input    Input
		expected []byte
	}{
		{input: Input{4, []uint64{1, 2, 3, 4}}, expected: []byte{0x21, 0x43}},
		{input: Input{6, []uint64{1, 2, 3, 4}}, expected: []byte{0x81, 0x30, 0x10}},
	}
	for ii, tt := range tests {
		t.Run(strconv.Itoa(ii), func(t *testing.T) {

			buf := &BitSetBuffer{}

			for _, i := range tt.input.values {
				WriteUint(buf, tt.input.numOfBits, binary.LittleEndian, i)
			}

			actual := buf.Bytes()

			if !reflect.DeepEqual(tt.expected, actual) {
				t.Fatalf("expected %v but found %v", tt.expected, actual)
			}
		})
	}
}

func TestBitSetBuffer_ReadBit(t *testing.T) {
	type fields struct {
		pos int
		Set []bool
	}
	tests := []struct {
		fields  fields
		wantBit bool
		wantErr bool
	}{
		{fields: fields{pos: 0, Set: []bool{}}, wantBit: false, wantErr: true},
		{fields: fields{pos: 0, Set: []bool{true}}, wantBit: true, wantErr: false},
		{fields: fields{pos: 1, Set: []bool{true}}, wantBit: false, wantErr: true},
	}
	for ii, tt := range tests {
		t.Run(strconv.Itoa(ii), func(t *testing.T) {
			bsb := &BitSetBuffer{
				pos: tt.fields.pos,
				Set: tt.fields.Set,
			}
			gotBit, err := bsb.ReadBit()
			if (err != nil) != tt.wantErr {
				t.Errorf("BitSetBuffer.ReadBit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotBit != tt.wantBit {
				t.Errorf("BitSetBuffer.ReadBit() = %v, want %v", gotBit, tt.wantBit)
			}
		})
	}
}

func TestWriteUintReadUint(t *testing.T) {

	values := make([]uint64, 100000)
	for i := range values {
		values[i] = rand.Uint64()
	}

	numOfBits := make([]int, len(values))
	for i := range values {
		numOfBits[i] = 64 - bits.LeadingZeros64(values[i])
	}

	type args struct {
		values     []uint64
		numOfBits  []int
		endianness binary.ByteOrder
	}
	tests := []struct {
		args    args
		wantErr bool
	}{
		{args{values, numOfBits, binary.LittleEndian}, false},
		{args{values, numOfBits, binary.BigEndian}, false},
	}
	for ii, tt := range tests {
		t.Run(strconv.Itoa(ii), func(t *testing.T) {
			buf := &BitSetBuffer{}

			if len(tt.args.values) != len(tt.args.numOfBits) {
				t.Fatalf("len(values) == len(numOfBits)")
			}
			var err error
			for i := range tt.args.values {
				err = WriteUint(buf, tt.args.numOfBits[i], tt.args.endianness, tt.args.values[i])
				if err != nil {
					t.Log(tt.args.values)
					t.Log(tt.args.numOfBits)
					t.Fatal(err)
				}
			}

			buf.ResetToStart()

			actual := make([]uint64, len(tt.args.numOfBits))
			for i := range actual {
				actual[i], err = ReadUint(buf, tt.args.numOfBits[i], tt.args.endianness)
				if err != nil {
					t.Fatal(err)
				}
			}

			if !reflect.DeepEqual(actual, tt.args.values) {
				t.Fatalf("expected: %v but found %v", tt.args.values, actual)
			}
		})
	}
}
