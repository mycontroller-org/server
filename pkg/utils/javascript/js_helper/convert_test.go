package javascript_helper

import (
	"testing"
)

func TestEndianConversions(t *testing.T) {
	conv := &Convert{}

	// test data
	leBytes16 := []byte{0x34, 0x12}                                     // 0x1234
	leBytes32 := []byte{0x78, 0x56, 0x34, 0x12}                         // 0x12345678
	leBytes64 := []byte{0xF0, 0xDE, 0xBC, 0x9A, 0x78, 0x56, 0x34, 0x12} // 0x123456789ABCDEF0

	beBytes16 := []byte{0x12, 0x34}
	beBytes32 := []byte{0x12, 0x34, 0x56, 0x78}
	beBytes64 := []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}

	t.Run("ToUInt16LE", func(t *testing.T) {
		want := uint16(0x1234)
		got := conv.ToUInt16LE(leBytes16)
		if got != want {
			t.Errorf("ToUInt16LE() = %d, want %d", got, want)
		}
	})

	t.Run("ToUInt32LE", func(t *testing.T) {
		want := uint32(0x12345678)
		got := conv.ToUInt32LE(leBytes32)
		if got != want {
			t.Errorf("ToUInt32LE() = %d, want %d", got, want)
		}
	})

	t.Run("ToUInt64LE", func(t *testing.T) {
		want := uint64(0x123456789ABCDEF0)
		got := conv.ToUInt64LE(leBytes64)
		if got != want {
			t.Errorf("ToUInt64LE() = %d, want %d", got, want)
		}
	})

	t.Run("ToUInt16BE", func(t *testing.T) {
		want := uint16(0x1234)
		got := conv.ToUInt16BE(beBytes16)
		if got != want {
			t.Errorf("ToUInt16BE() = %d, want %d", got, want)
		}
	})

	t.Run("ToUInt32BE", func(t *testing.T) {
		want := uint32(0x12345678)
		got := conv.ToUInt32BE(beBytes32)
		if got != want {
			t.Errorf("ToUInt32BE() = %d, want %d", got, want)
		}
	})

	t.Run("ToUInt64BE", func(t *testing.T) {
		want := uint64(0x123456789ABCDEF0)
		got := conv.ToUInt64BE(beBytes64)
		if got != want {
			t.Errorf("ToUInt64BE() = %d, want %d", got, want)
		}
	})

	t.Run("ToInt16LE", func(t *testing.T) {
		bytes := []byte{0xFF, 0x7F} // 0x7FFF -> 32767
		want := int16(32767)
		got := conv.ToInt16LE(bytes)
		if got != want {
			t.Errorf("ToInt16LE() = %d, want %d", got, want)
		}

		bytes = []byte{0x00, 0x80} // 0x8000 -> -32768
		want = -32768
		got = conv.ToInt16LE(bytes)
		if got != want {
			t.Errorf("ToInt16LE() = %d, want %d", got, want)
		}
	})

	t.Run("ToInt32BE", func(t *testing.T) {
		bytes := []byte{0xFF, 0xFF, 0xFF, 0xFF} // -1
		want := int32(-1)
		got := conv.ToInt32BE(bytes)
		if got != want {
			t.Errorf("ToInt32BE() = %d, want %d", got, want)
		}
	})
}
