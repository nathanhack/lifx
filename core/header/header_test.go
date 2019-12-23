package header

import (
	"fmt"
	"testing"
)

func TestHeader_SetOrigin(t *testing.T) {
	h := Header{}
	if h.Origin() != 0 {
		t.Errorf("expect 0 but got %v", h.Origin())
	}
	h.SetOrigin(3)
	if h.Origin() != 3 {
		t.Errorf("expect 3 but got %v", h.Origin())
	}
	if h[3] != 0xc0 {
		t.Errorf("expect 0xc0 but got 0x%02x", h[3])
	}
}

func TestHeader_Tagged(t *testing.T) {
	h := Header{}
	if h.Tagged() {
		t.Errorf("expect false but got %v", h.Tagged())
	}
	h.SetTagged(true)
	if !h.Tagged() {
		t.Errorf("expect true but got %v", h.Tagged())
	}
	if h[3] != 0x20 {
		t.Errorf("expect 0x20 but got 0x%02x", h[3])
	}
	fmt.Printf("%x", h)
}

func TestHeader_Addressable(t *testing.T) {
	h := Header{}
	if h.Addressable() {
		t.Errorf("expect false but got %v", h.Addressable())
	}
	h.SetAddressable(true)
	if !h.Addressable() {
		t.Errorf("expect true but got %v", h.Addressable())
	}
	if h[3] != 0x20 {
		t.Errorf("expect 0x20 but got 0x%02x", h[3])
	}
	fmt.Printf("%x", h)
}
