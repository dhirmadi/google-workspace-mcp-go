package color

import "strings"

// HexToRGB converts a hex color string (#RRGGBB or RRGGBB) to RGB float64
// values in the range [0.0, 1.0]. Returns (0,0,0, false) if the input is invalid.
func HexToRGB(hex string) (r, g, b float64, ok bool) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, 0, false
	}
	return float64(hexToByte(hex[0:2])) / 255.0,
		float64(hexToByte(hex[2:4])) / 255.0,
		float64(hexToByte(hex[4:6])) / 255.0,
		true
}

// hexToByte converts a 2-char hex string to a byte value.
func hexToByte(hex string) byte {
	var val byte
	for _, c := range hex {
		val *= 16
		switch {
		case c >= '0' && c <= '9':
			val += byte(c - '0')
		case c >= 'a' && c <= 'f':
			val += byte(c-'a') + 10
		case c >= 'A' && c <= 'F':
			val += byte(c-'A') + 10
		}
	}
	return val
}
