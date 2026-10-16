package mysensors

import "testing"

func TestDetectBootloaderKind(t *testing.T) {
	tests := []struct {
		name string
		hex  string
		want string
	}{
		{"MYSBootloader 1.0 (6 bytes)", "E70019004802", BootloaderKindMYSBootloader10},
		{"MYSBootloader 1.3", "E700190048025B0E0103", BootloaderKindMYSBootloader13},
		{"MYSBootloader 1.4", "E700190048025B0E0104", BootloaderKindMYSBootloader14},
		{"Protocol 3.1 (11 bytes)", "E700190048025B0E010310", BootloaderKindProtocol31},
		{"Protocol 3.1 extended (18 bytes)", "E700190048025B0E010310A10000000000AB000000", BootloaderKindProtocol31},
		{"unknown 10-byte", "E700190048025B0E00FF", BootloaderKindUnknown},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectBootloaderKind(tt.hex)
			if got != tt.want {
				t.Errorf("detectBootloaderKind(%q) = %q, want %q", tt.hex, got, tt.want)
			}
		})
	}
}

func TestExtractBootloaderVersion(t *testing.T) {
	tests := []struct {
		name string
		hex  string
		want string
	}{
		{"1.3", "E700190048025B0E0103", "1.3"},
		{"1.4", "E700190048025B0E0104", "1.4"},
		{"6 bytes no version", "E70019004802", ""},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractBootloaderVersion(tt.hex)
			if got != tt.want {
				t.Errorf("extractBootloaderVersion(%q) = %q, want %q", tt.hex, got, tt.want)
			}
		})
	}
}

func TestIsMYSBootloaderKind(t *testing.T) {
	if !isMYSBootloaderKind("mysbootloader-1.4") {
		t.Fatal("1.4 must be MYSBootloader")
	}
	if !isMYSBootloaderKind("mysbootloader-1.0") {
		t.Fatal("1.0 must be MYSBootloader")
	}
	if isMYSBootloaderKind("protocol-3.1") {
		t.Fatal("protocol-3.1 must not be MYSBootloader")
	}
	if isMYSBootloaderKind("unknown") {
		t.Fatal("unknown must not be MYSBootloader")
	}
}
