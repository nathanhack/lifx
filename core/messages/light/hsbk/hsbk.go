package hsbk

import "fmt"

type HSBK struct {
	Hue        uint16 //Hue: range 0 to 65535 (scaled 0 to 360)
	Saturation uint16 //Saturation: range 0 to 65535 (scaled 0% to 100%)
	Brightness uint16 //Brightness: range 0 to 65535 (scaled 0% to 100%)
	Kelvin     uint16 //Kelvin: range 2500° (warm) to 9000° (cool)
}

func (hsbk HSBK) ToRGB() (r, g, b float32) {
	h := float32(hsbk.Hue) / 0xffff * 360
	s := float32(hsbk.Saturation) / 0xffff * 100
	l := float32(hsbk.Brightness) / 0xffff * 100

	if s == 0 {
		return 0, 0, 0
	}

	var v1, v2 float32
	if l < 0.5 {
		v2 = l * (1 + s)
	} else {
		v2 = (l + s) - (s * l)
	}

	v1 = 2*l - v2

	r = hueToRGB(v1, v2, h+(1.0/3.0))
	g = hueToRGB(v1, v2, h)
	b = hueToRGB(v1, v2, h-(1.0/3.0))

	return
}

func (hsbk HSBK) String() string {
	h := float32(hsbk.Hue) / 0xffff * 360
	s := float32(hsbk.Saturation) / 0xffff * 100
	l := float32(hsbk.Brightness) / 0xffff * 100
	return fmt.Sprintf("Hue:%.2f Sat:%.2f%% Bright:%.2f%% Kelvin:%v", h, s, l, hsbk.Kelvin)
}

func hueToRGB(v1, v2, h float32) float32 {
	if h < 0 {
		h += 1
	}
	if h > 1 {
		h -= 1
	}
	switch {
	case 6*h < 1:
		return (v1 + (v2-v1)*6*h)
	case 2*h < 1:
		return v2
	case 3*h < 2:
		return v1 + (v2-v1)*((2.0/3.0)-h)*6
	}
	return v1
}
