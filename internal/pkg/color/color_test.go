package color

import (
	"math"
	"testing"
)

func TestHexToRGB(t *testing.T) {
	tests := []struct {
		name       string
		hex        string
		r, g, b    float64
		wantOK     bool
	}{
		{"black", "#000000", 0, 0, 0, true},
		{"white", "#FFFFFF", 1, 1, 1, true},
		{"red", "#FF0000", 1, 0, 0, true},
		{"green", "#00FF00", 0, 1, 0, true},
		{"blue", "#0000FF", 0, 0, 1, true},
		{"no hash", "FF8800", 1, 0x88 / 255.0, 0, true},
		{"lowercase", "#ff8800", 1, 0x88 / 255.0, 0, true},
		{"mixed case", "#Ff8800", 1, 0x88 / 255.0, 0, true},
		{"too short", "#FFF", 0, 0, 0, false},
		{"too long", "#FFFFFFFF", 0, 0, 0, false},
		{"empty", "", 0, 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b, ok := HexToRGB(tt.hex)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if math.Abs(r-tt.r) > 0.002 || math.Abs(g-tt.g) > 0.002 || math.Abs(b-tt.b) > 0.002 {
				t.Errorf("got (%f, %f, %f), want (%f, %f, %f)", r, g, b, tt.r, tt.g, tt.b)
			}
		})
	}
}
