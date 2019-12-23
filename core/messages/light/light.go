package light

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/nathanhack/lifx/core/header"
	"github.com/nathanhack/lifx/core/messages/light/hsbk"
)

const (
	GetType             = 101
	SetColorType        = 102
	SetWaveformType     = 103
	SetWaveformOptional = 119
	StateType           = 107
	GetPowerType        = 116
	SetPowerType        = 117
	StatePowerType      = 118
	GetInfraredType     = 120
	StateInfraredType   = 121
	SetInfraredType     = 122
)

type Get [0]byte

func (g Get) RequiredHeader(h *header.Header) {
	h.SetType(GetType)
	h.SetResponseRequired(true)
	h.SetSize(header.HeaderLen)
}

type SetColor struct {
	Reserved uint8
	Color    hsbk.HSBK
	Duration uint32 // transition time in milliseconds
}

func (g SetColor) RequiredHeader(h *header.Header, responseRequired bool) {
	h.SetType(SetColorType)
	// if true a response with state will be sent
	h.SetResponseRequired(responseRequired)
	h.SetSize(header.HeaderLen)
}

type State struct {
	Color     hsbk.HSBK
	Reserved1 int16
	Power     uint16
	Label     [32]byte
	Reseved2  uint64
}

func (s State) GetPower() bool {
	return s.Power == 0xffff
}

func (s State) GetLabel() string {
	return string(s.Label[:])
}

func (s State) String() string {
	if s.GetPower() {
		return fmt.Sprintf("Color:{%v} Power:ON Label:%v", s.Color, string(s.Label[:]))
	}
	return fmt.Sprintf("Color:{%v} Power:OFF Label:%v", s.Color, string(s.Label[:]))
}

type GetPower [0]byte

func (g GetPower) RequiredHeader(h *header.Header) {
	h.SetType(GetPowerType)
	h.SetResponseRequired(true)
	h.SetSize(header.HeaderLen)
}

type SetPower struct {
	Level    uint16
	Duration uint32 // transition time in milliseconds
}

func (SetPower) RequiredHeader(h *header.Header, responseRequired bool) {
	h.SetType(SetPowerType)
	// if true a response with state will be sent
	h.SetResponseRequired(responseRequired)
	h.SetSize(header.HeaderLen)
}

func (sp SetPower) GetLevel() bool {
	return sp.Level == 0xffff
}

func (sp *SetPower) SetLevel(on bool) {
	if on {
		sp.Level = 0xffff
	} else {
		sp.Level = 0
	}
}

type StatePower struct {
	Level uint16
}

func (sp StatePower) GetLevel() bool {
	return sp.Level == 0xffff
}
func (sp *StatePower) SetLevel(on bool) {
	if on {
		sp.Level = 0xffff
	} else {
		sp.Level = 0
	}
}

func (sp StatePower) String() string {
	if sp.GetLevel() {
		return "{Level:ON}"
	}
	return "{Level:OFF}"
}

type GetInfrared [0]byte

func (g GetInfrared) RequiredHeader(h *header.Header) {
	h.SetType(GetInfraredType)
	h.SetResponseRequired(true)
	h.SetSize(header.HeaderLen)
}

type SetInfrared struct {
	Brightness uint16
}

func (g SetInfrared) RequiredHeader(h *header.Header, responseRequired bool) {
	h.SetType(SetInfraredType)
	h.SetResponseRequired(responseRequired)
	h.SetSize(header.HeaderLen)
}

type StateInfrared struct {
	Brightness uint16
}

func DecodeFromHeader(h header.Header, message interface{}) error {
	switch h.Type() {
	case StateType:
		if s, ok := message.(*State); ok {
			buffer := bytes.NewBuffer(h.Data())
			err := binary.Read(buffer, binary.LittleEndian, s)
			if err != nil {
				return err
			}
		}
	case StatePowerType:
		if s, ok := message.(*StatePower); ok {
			buffer := bytes.NewBuffer(h.Data())
			err := binary.Read(buffer, binary.LittleEndian, s)
			if err != nil {
				return err
			}
		}
	case StateInfraredType:
		if s, ok := message.(*StateInfrared); ok {
			buffer := bytes.NewBuffer(h.Data())
			err := binary.Read(buffer, binary.LittleEndian, s)
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("not supported")
	}
	return nil
}
