package ssz

import (
	"bytes"
	"testing"
)
func TestEncodeUint32(t *testing.T) {
	buf := make([]byte, 100)
	EncodeUint32(buf, 0, uint32(0x3456789a))
	exp := []byte{0x34, 0x56, 0x78, 0x9a}
	if bytes.Compare(buf[:4], exp) != 0 {
		t.Errorf("got %x, expected %x", buf[:4], exp)
	}
}

func TestEncodeUint256(t *testing.T) {
	buf := make([]byte, 100)
	offset, _ := EncodeUintN(buf, 0, 256, big.NewInt(0x3456789a))
	exp := []byte{0x34, 0x56, 0x78, 0x9a}
	if bytes.Compare(buf[:4], exp) != 0 {
		t.Errorf("got %x, expected %x", buf[:4], exp)
	}
	if offset != 256 {
		t.Errorf("Offset wrong, expected 256, got %d", offset)
	}
}
