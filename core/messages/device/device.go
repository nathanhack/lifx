package device

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/nathanhack/lifx/core/header"
)

const (
	GetServiceType        = 2
	StateServiceType      = 3
	GetHostInfoType       = 12
	StateHostInfoType     = 13
	GetHostFirmwareType   = 14
	StateHostFirmwareType = 15
	GetWifiInfoType       = 16
	StateWifiInfoType     = 17
	GetWifiFirmwareType   = 18
	StateWifiFirmwareType = 19
	GetPowerType          = 20
	SetPowerType          = 21
	StatePowerType        = 22
	GetLabelType          = 23
	SetLabelType          = 24
	StateLabelType        = 25
	GetVersionType        = 32
	StateVersionType      = 33
	GetInfoType           = 34
	StateInfoType         = 35
	AcknowledgementType   = 45
	GetLocationType       = 48
	SetLocationType       = 49
	StateLocationType     = 50
	GetGroupType          = 51
	SetGroupType          = 52
	StateGroupType        = 53
	EchoRequestType       = 58
	EchoResponseType      = 59
)

type GetService [0]byte

func (g GetService) GetBytes() []byte {
	return []byte{}
}

func (g GetService) RequiredHeader(h *header.Header) {
	h.SetTagged(true)
	h.SetType(GetServiceType)
	h.SetResponseRequired(true)
	h.SetSize(header.HeaderLen)
}

type StateService struct {
	Service byte   //maps to Service (https://lan.developer.lifx.com/v2.0/docs/device-messages#section-service) should always be 0x1
	Port    uint32 // this port should when sending messages
}

type GetHostInfo [0]byte

func (g GetHostInfo) RequiredHeader(h *header.Header) {
	h.SetType(GetHostInfoType)
	h.SetSize(header.HeaderLen)
}

type StateHostInfo struct {
	Signal   float32
	Tx       uint32
	Rx       uint32
	reserved int16
}

type GetHostFirmware [0]byte

func (g GetHostFirmware) RequiredHeader(h *header.Header) {
	h.SetType(GetHostFirmwareType)
	h.SetSize(header.HeaderLen)
}

type StateHostFirmware struct {
	Build        uint64
	reserved     uint64
	VersionMinor uint64
	VersionMajor uint64
}
type GetWifiInfo [0]byte

func (g GetWifiInfo) RequiredHeader(h *header.Header) {
	h.SetType(GetWifiInfoType)
	h.SetSize(header.HeaderLen)
}

type WifiStrength string

const (
	NoSignal    WifiStrength = "No signal"
	VeryBad     WifiStrength = "Very bad signal"
	SomewhatBad WifiStrength = "Somewhat bad signal"
	Alright     WifiStrength = "Alright signal"
	Good        WifiStrength = "Good signal"
)

type StateWifiInfo struct {
	Signal   float32
	Tx       uint32
	Rx       uint32
	reserved int16
}

func (s StateWifiInfo) SignalInfo() WifiStrength {
	if s.Signal < 0 || s.Signal == 200 {
		switch {
		case s.Signal == 200:
			return NoSignal
		case s.Signal <= -80:
			return VeryBad
		case s.Signal <= -70:
			return SomewhatBad
		case s.Signal <= -60:
			return Alright
		default:
			return Good
		}
	} else {
		switch {
		case s.Signal == 4 || s.Signal == 5:
			return VeryBad
		case s.Signal >= 7 && s.Signal <= 11:
			return SomewhatBad
		case s.Signal >= 12 && s.Signal <= 16:
			return Alright
		case s.Signal > 16:
			return Good
		default:
			return NoSignal
		}
	}
}

type GetWifiFirmware [0]byte

func (g GetWifiFirmware) RequiredHeader(h *header.Header) {
	h.SetType(GetWifiFirmwareType)
	h.SetSize(header.HeaderLen)
}

type StateWifiFirmware struct {
}
type GetPower [0]byte

func (g GetPower) RequiredHeader(h *header.Header) {
	h.SetType(GetPowerType)
	h.SetSize(header.HeaderLen)
}

type SetPower struct {
	Level uint16
}

func (s SetPower) GetLevel() bool {
	return s.Level > 0
}
func (s *SetPower) SetLevel(on bool) {
	if on {
		s.Level = 0xffff
	} else {
		s.Level = 0
	}
}

func (s SetPower) RequiredHeader(h *header.Header, responseRequired bool) {
	h.SetType(SetPowerType)
	h.SetResponseRequired(responseRequired)
	h.SetSize(header.HeaderLen + 2)
}

type StatePower struct {
	Level uint16
}

func (s *StatePower) SetLevel(on bool) {
	if on {
		s.Level = 0xffff
	} else {
		s.Level = 0
	}
}

func (s *StatePower) GetLevel() bool {
	return s.Level == 0xffff
}

type GetLabel [0]byte

func (g GetLabel) RequiredHeader(h *header.Header) {
	h.SetType(GetLabelType)
	h.SetSize(header.HeaderLen)
}

type SetLabel struct {
	Label [32]byte //string
}

type StateLabel struct {
	Label [32]byte //string
}

func (l StateLabel) GetLabel() string {
	return string(l.Label[:])
}

func (l StateLabel) String() string {
	return fmt.Sprintf("StateLabel{Label:'%v'}", l.GetLabel())
}

type GetVersion [0]byte

func (g GetVersion) RequiredHeader(h *header.Header) {
	h.SetType(GetVersionType)
	h.SetSize(header.HeaderLen)
}

type StateVersion struct {
	Vendor  uint32
	Product uint32
	Version uint32
}

type GetInfo [0]byte

func (g GetInfo) RequiredHeader(h *header.Header) {
	h.SetType(GetInfoType)
	h.SetSize(header.HeaderLen)
}

type StateInfo struct {
	Time     uint64
	Uptime   uint64
	Downtime uint64
}
type Acknowledgement [0]byte

type GetLocation [0]byte

func (g GetLocation) RequiredHeader(h *header.Header) {
	h.SetType(GetLocationType)
	h.SetSize(header.HeaderLen)
}

type SetLocation struct {
	Location  [16]byte
	Label     [32]byte //string
	UpdatedAt uint64
}
type StateLocation struct {
	Location  [16]byte
	Label     [32]byte //string
	UpdatedAt uint64
}

type GetGroup [0]byte

func (g GetGroup) RequiredHeader(h *header.Header) {
	h.SetType(GetGroupType)
	h.SetSize(header.HeaderLen)
}

type SetGroup struct {
	Group     [16]byte
	Label     [32]byte //string
	UpdatedAt uint64
}

type StateGroup struct {
	Group     [16]byte
	Label     [32]byte //string
	UpdatedAt uint64
}

type EchoRequest [64]byte

type EchoResponse [64]byte

func DecodeFromHeader(h header.Header, message interface{}) error {
	switch h.Type() {
	case StateServiceType:
		if s, ok := message.(*StateService); ok {
			buffer := bytes.NewBuffer(h.Data())
			err := binary.Read(buffer, binary.LittleEndian, s)
			if err != nil {
				return err
			}
		}
	case StateHostInfoType:
		if s, ok := message.(*StateHostInfo); ok {
			buffer := bytes.NewBuffer(h.Data())
			err := binary.Read(buffer, binary.LittleEndian, s)
			if err != nil {
				return err
			}
		}
	case StateHostFirmwareType:
		if s, ok := message.(*StateHostFirmware); ok {
			buffer := bytes.NewBuffer(h.Data())
			err := binary.Read(buffer, binary.LittleEndian, s)
			if err != nil {
				return err
			}
		}
	case StateWifiInfoType:
		if s, ok := message.(*StateWifiInfo); ok {
			buffer := bytes.NewBuffer(h.Data())
			err := binary.Read(buffer, binary.LittleEndian, s)
			if err != nil {
				return err
			}
		}
	case StateWifiFirmwareType:
		if s, ok := message.(*StateWifiFirmware); ok {
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
	case StateLabelType:
		if s, ok := message.(*StateLabel); ok {
			buffer := bytes.NewBuffer(h.Data())
			err := binary.Read(buffer, binary.LittleEndian, s)
			if err != nil {
				return err
			}
		}
	case StateVersionType:
		if s, ok := message.(*StateVersion); ok {
			buffer := bytes.NewBuffer(h.Data())
			err := binary.Read(buffer, binary.LittleEndian, s)
			if err != nil {
				return err
			}
		}
	case StateInfoType:
		if s, ok := message.(*StateInfo); ok {
			buffer := bytes.NewBuffer(h.Data())
			err := binary.Read(buffer, binary.LittleEndian, s)
			if err != nil {
				return err
			}
		}
	case StateLocationType:
		if s, ok := message.(*StateLocation); ok {
			buffer := bytes.NewBuffer(h.Data())
			err := binary.Read(buffer, binary.LittleEndian, s)
			if err != nil {
				return err
			}
		}
	case StateGroupType:
		if s, ok := message.(*StateGroup); ok {
			buffer := bytes.NewBuffer(h.Data())
			err := binary.Read(buffer, binary.LittleEndian, s)
			if err != nil {
				return err
			}
		}
	case EchoResponseType:
		if s, ok := message.(*EchoResponse); ok {
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
