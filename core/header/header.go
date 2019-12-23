package header

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"reflect"
)

const HeaderLen = 36

type Header []byte

func Decode(bytes []byte) (*Header, error) {
	if len(bytes) < HeaderLen {
		return nil, fmt.Errorf("expected %v or more bytes found %v", HeaderLen, len(bytes))
	}

	h := Header(bytes)
	return &h, nil
}

func New() *Header {
	head := Header(make([]byte, HeaderLen))
	head.SetProtocol(1024)    //must always
	head.SetAddressable(true) //must always
	head.SetSource(rand.Uint32())
	return &head
}

func (h Header) Size() uint16 {
	return uint16(h[0]) + (uint16(h[1]))<<8
}

func (h Header) SetSize(size uint16) {
	h[0] = byte(size & 0xff)
	h[1] = byte((size & 0xff00) >> 8)
}

func (h Header) Protocol() uint16 {
	return uint16(h[2]) + (uint16(h[3])&0xf)<<8
}

func (h Header) SetProtocol(protocol uint16) {
	h[2] = byte(protocol & 0xff)
	h[3] = byte((protocol & 0x0f00) >> 8)
}

func (h Header) Addressable() bool {
	return h[3]&(1<<4) > 0
}

func (h Header) SetAddressable(addressable bool) {
	if addressable {
		h[3] |= 1 << 4
	} else {
		h[3] = h[3] &^ (1 << 4)
	}
}

func (h Header) Tagged() bool {
	return h[3]&(1<<5) > 0
}

func (h Header) SetTagged(tagged bool) {
	if tagged {
		h[3] |= 1 << 5
	} else {
		h[3] = h[3] &^ (1 << 5)
	}
}

func (h Header) Origin() byte {
	//origin must aways be zero
	return h[3] >> 6
}

func (h Header) SetOrigin(origin byte) {
	h[3] |= (origin & 0b11) << 6
}

func (h Header) Source() uint32 {
	return uint32(h[4]) + uint32(h[5])<<8 + uint32(h[6])<<16 + uint32(h[7])<<24
}

func (h Header) SetSource(source uint32) {
	h[4] = byte((source & 0xff) >> 0)
	h[5] = byte((source & 0xff00) >> 8)
	h[6] = byte((source & 0xff0000) >> 16)
	h[7] = byte((source & 0xff000000) >> 24)
}

func (h Header) Target() []byte {
	return h[8:15]
}

func (h Header) TargetHex() string {
	return hex.EncodeToString(h.Target())
}

func (h Header) SetTarget(target []byte) {
	copy(h[8:15], target)
}

func (h Header) FrameAddressReservedReset() {
	copy(h[16:21], []byte{0, 0, 0, 0, 0, 0})
}

func (h Header) ResponseRequired() bool {
	return h[22]&0x01 > 0
}

func (h Header) SetResponseRequired(required bool) {
	if required {
		h[22] |= 0b1
	} else {
		h[22] = h[22] &^ 0b1
	}
}

func (h Header) AcknowledgementRequired() bool {
	return h[22]&0b10 > 0
}

func (h Header) SetAcknowledgementRequired(required bool) {
	if required {
		h[22] |= 0b10
	} else {
		h[22] = h[22] &^ 0b10
	}
}

func (h Header) Sequence() byte {
	return h[23]
}

func (h Header) SetSequence(seq byte) {
	h[23] = seq
}

func (h Header) Type() uint16 {
	return uint16(h[32]) + uint16(h[33])<<8
}

func (h Header) SetType(packetType uint16) {
	h[32] = byte(packetType & 0xff)
	h[33] = byte((packetType & 0xff00) >> 8)
}

func (h Header) Validate(filterSenders bool) bool {
	//weak but better than nothing
	l := h[16:22]
	zeros := reflect.DeepEqual([]byte{0, 0, 0, 0, 0, 0}, l)
	s := string(l)
	return len(s) == 6 &&
		(zeros || s == "LIFXV2") && (!filterSenders || !zeros)
}

func (h Header) Data() []byte {
	return h[HeaderLen:]
}

func (h Header) String() string {
	return fmt.Sprintf("{ Size:%v Protocol:%v Addressable:%v Tagged:%v Origin:%v Source:%v Target:%x Response:%v Acknowledgment:%v Sequence:%v Type:%v }",
		h.Size(),
		h.Protocol(),
		h.Addressable(),
		h.Tagged(),
		h.Origin(),
		h.Source(),
		h.Target(),
		h.ResponseRequired(),
		h.AcknowledgementRequired(),
		h.Sequence(),
		h.Type(),
	)
}
