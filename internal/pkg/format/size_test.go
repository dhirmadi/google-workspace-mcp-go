package format

import "testing"

func TestByteSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, ""},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{5242880, "5.0 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := ByteSize(tt.bytes)
			if got != tt.want {
				t.Errorf("ByteSize(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}
