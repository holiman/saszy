package ssz

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"reflect"
)

func getUint24(b []byte) uint32 {
	_ = b[2] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16
}
func DecodeUint8(buf []byte, offset uint32) (uint8, uint32) {
	return uint8(buf[offset]), offset + 1
}

func DecodeUint16(buf []byte, offset uint32) (uint16, uint32) {
	return binary.BigEndian.Uint16(buf[offset:]), offset + 2
}

func DecodeUint24(buf []byte, offset uint32) (uint32, uint32) {
	return getUint24(buf[offset:]), offset + 3
}

func DecodeUint32(buf []byte, offset uint32) (uint32, uint32) {
	return binary.BigEndian.Uint32(buf[offset:]), offset + 4
}

func DecodeUint64(buf []byte, offset uint32) (uint64, uint32) {
	return binary.BigEndian.Uint64(buf[offset:]), offset + 8
}

func DecodeUintN(buf []byte, offset, n uint32, dest *big.Int) (*big.Int, uint32, error) {
	if n%8 != 0 {
		// error
		return nil, 0, fmt.Errorf("invalid 'uintn' format: uint%d not a multiple of '8'", n)
	}
	if dest == nil {
		dest = new(big.Int)
	}
	dest.SetBytes(buf[offset : offset+n/8])
	return dest, offset + n/8, nil
}

func DecodeBool(buf []byte, offset uint32) (bool, uint32, error) {
	if buf[offset] == 0 {
		return false, offset + 1, nil
	}
	if buf[offset] == 1 {
		return true, offset + 1, nil
	}
	return true, offset + 1, fmt.Errorf("invalid boolean value %d", buf[offset])
}

func DecodeBytesX(buf []byte, offset uint32) ([]byte, uint32, error) {
	bytelen, offset := DecodeUint32(buf, offset)
	dest := make([]byte, bytelen)
	copy(dest, buf[offset:offset+bytelen])
	return dest, offset + bytelen, nil
}

func DecodeBytesN(buf []byte, offset, length uint32, dest []byte) (uint32, error) {
	copy(dest, buf[offset:offset+length])
	return offset + length, nil
}

func DecodeListUint32(buf []byte, offset, elemSize uint32) ([]uint32, uint32, error) {
	bytelen, offset := DecodeUint32(buf, offset)
	if uint32(len(buf)) < offset+bytelen {
		return []uint32{}, 0, fmt.Errorf("insufficient input for list")
	}
	nElems := bytelen / elemSize
	retval := make([]uint32, nElems)
	for i := uint32(0); i < nElems; i++ {
		retval[i], offset = DecodeUint32(buf, offset)
	}
	return retval, offset, nil
}

// helper method, inspired from the BigEndian lib, which does not support 24-bit ints
func putUint24(b []byte, v uint32) {
	_ = b[3] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
}

func EncodeUint8(buf []byte, offset uint32, val uint8) (uint32, error) {
	if req := offset + 1; req < uint32(len(buf)) {
		return 0, fmt.Errorf("insufficient buffer space, have %d require %d", len(buf), req)
	}
	buf[offset] = byte(val)
	return offset + 1, nil
}
func EncodeUint16(buf []byte, offset uint32, val uint16) (uint32, error) {
	if req := offset + 2; req < uint32(len(buf)) {
		return 0, fmt.Errorf("insufficient buffer space, have %d require %d", len(buf), req)
	}
	binary.BigEndian.PutUint16(buf[offset:], val)
	return offset + 2, nil
}
func EncodeUint24(buf []byte, offset uint32, val uint32) (uint32, error) {
	if req := offset + 3; req < uint32(len(buf)) {
		return 0, fmt.Errorf("insufficient buffer space, have %d require %d", len(buf), req)
	}
	putUint24(buf[offset:], val)
	return offset + 3, nil
}
func EncodeUint32(buf []byte, offset uint32, val uint32) (uint32, error) {
	if req := offset + 4; req < uint32(len(buf)) {
		return 0, fmt.Errorf("insufficient buffer space, have %d require %d", len(buf), req)
	}
	binary.BigEndian.PutUint32(buf[offset:], val)
	return offset + 4, nil
}

func EncodeUint64(buf []byte, offset uint32, val uint64) (uint32, error) {
	if req := offset + 8; req < uint32(len(buf)) {
		return 0, fmt.Errorf("insufficient buffer space, have %d require %d", len(buf), req)
	}
	binary.BigEndian.PutUint64(buf[offset:], val)
	return offset + 8, nil
}
func EncodeUintN(buf []byte, offset, n uint32, val *big.Int) (uint32, error) {
	if n%8 != 0 {
		// error
		return 0, fmt.Errorf("invalid 'uintn' format: uint%d not a multiple of '8'", n)
	}
	bigintBytes := val.Bytes()
	numBytes := n / 8
	if diff := numBytes - uint32(len(bigintBytes)); diff >= 0 {
		// We might want to zero-fill here
		copy(buf[offset+diff:], bigintBytes)
	} else {
		// Now, we have an integer which does not fit in the required space.
		// Just trim upper bytes, probably the least wrong thing to do
		copy(buf[offset:], bigintBytes[-diff:])
	}
	return offset + numBytes, nil
}

func EncodeBool(buf []byte, offset uint32, val bool) (uint32, error) {
	if req := offset + 1; req < uint32(len(buf)) {
		return 0, fmt.Errorf("insufficient buffer space, have %d require %d", len(buf), req)
	}
	if val {
		buf[offset] = 1
	} else {
		buf[offset] = 0
	}
	return offset + 1, nil
}

// EncodeBytesWithoutLengthPrefix encodes a fixed length buffer
// which is not prefixed by bytelength
func EncodeBytesWithoutLengthPrefix(buf []byte, offset uint32, data []byte) (uint32, error) {
	if req := offset + uint32(len(data)); req < uint32(len(buf)) {
		return 0, fmt.Errorf("insufficient buffer space, have %d require %d", len(buf), req)
	}
	copy(buf[offset:], data)
	return offset + uint32(len(buf)), nil
}

// EncodeBytesWithLengthPrefix encodes a dynamic size buffer, [ len (uint32)  .. data ]
func EncodeBytesWithLengthPrefix(buf []byte, offset uint32, data []byte) (uint32, error) {
	nBytes := uint32(len(data))
	if req := offset + nBytes; req < uint32(len(buf)) {
		return 0, fmt.Errorf("insufficient buffer space, have %d require %d", len(buf), req)
	}
	offset, _ = EncodeUint32(buf, offset, nBytes)
	copy(buf[offset:], data)
	return offset + nBytes, nil
}


// SszTypeSize return the number of bytes required to encode the basic types.
// If the type is a non-fixed size, the method returns an error
// If the type is not an ssz type at all, the method returns an erorr
func SszTypeSize(sszType string) (int, error) {

	switch sszType {
	case "uint8", "bytes1":
		return 1, nil
	case "uint16", "bytes2":
		return 2, nil
	case "uint32", "bytes4":
		return 4, nil
	case "uint64", "bytes8":
		return 8, nil
	case "bytes20":
		return 20, nil
	case "uint256", "bytes32":
		return 32, nil
	}
	return 0, fmt.Errorf("ssz type %v is either dynamic-size or not a valid type", sszType)
}

func SszSize(sszObj SszObject) uint32 {
	if sszObj == nil || (reflect.ValueOf(sszObj).Kind() == reflect.Ptr && reflect.ValueOf(sszObj).IsNil()) {
		return 0
	}
	return sszObj.SszSize()
}

// SszEncode encodes an SszObject into the given buffer. It handles the
// case of 'nil' SszObject
func SszEncode(buf []byte, sszObj SszObject) (uint32, error) {

	size := SszSize(sszObj)
	offset, err := EncodeUint32(buf, 0, size)
	if err != nil{
		return offset, err
	}
	if size > 0 { // means object is non-nil
		return sszObj.EncodeSSZ(buf[offset:])
	}
	return offset, nil
}

type SszObject interface {
	SszSize() uint32
	EncodeSSZ([]byte) (uint32, error)
	DecodeSSZ([]byte) error
}
