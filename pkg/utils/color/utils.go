package color

import (
	"strconv"
)

// HexToFloat32 converts hex value into float32
func HexToFloat32(value string) float32 {
	val, err := strconv.ParseInt(value, 16, 64)
	if err != nil {
		return 0
	}
	return float32(val)
}

func hueToRGB(v1, v2, h float64) float64 {
	if h < 0 {
		h += 1
	}
	if h > 1 {
		h -= 1
	}
	switch {
	case h < 1.0/6.0:
		return (v1 + (v2-v1)*6.0*h)
	case h < 0.5:
		return v2
	case h < 2.0/3.0:
		return v1 + (v2-v1)*((2.0/3.0)-h)*6.0
	}
	return v1
}

// ToRGB returns Red, Green, Blue
// input hue, saturation, lightness
// hue range: 0~360
// saturation: 0.0~99
// lightness: 0.0~99
func ToRGB(h, s, l float64) (float32, float32, float32) {
	hue := h / 360.0
	saturation := s / 100.0
	lightness := l / 100.0

	if saturation == 0 {
		// it's gray
		return float32(l), float32(l), float32(l)
	}

	var v1, v2 float64
	if lightness < 0.5 {
		v2 = lightness * (1 + saturation)
	} else {
		v2 = (lightness + saturation) - (lightness * saturation)
	}

	v1 = 2.0*lightness - v2

	r := hueToRGB(v1, v2, hue+(1.0/3.0))
	g := hueToRGB(v1, v2, hue)
	b := hueToRGB(v1, v2, hue-(1.0/3.0))

	return float32(r * 255.0), float32(g * 255.0), float32(b * 255)
}
