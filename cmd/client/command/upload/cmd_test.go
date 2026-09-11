package upload

import (
	"bytes"
	"strings"
	"testing"
	"time"

	firmwareTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	"github.com/stretchr/testify/assert"
)

func TestFormatFileSize(t *testing.T) {
	assert.Equal(t, "0 B", formatFileSize(0))
	assert.Equal(t, "512 B", formatFileSize(512))
	assert.Equal(t, "67.83 KiB", formatFileSize(69458))
	assert.Equal(t, "1.50 MiB", formatFileSize(1572864))
}

func TestPrintFirmwareUpload(t *testing.T) {
	out := &bytes.Buffer{}
	printFirmwareUpload(out, &firmwareTY.Firmware{
		ID: "stm32-app-slot-a",
		File: firmwareTY.FileConfig{
			Name:         "firmware.signed.bin",
			InternalName: "stm32-app-slot-a.bin",
			Checksum:     "sha256:b1f3c09e4acbbf8ca40e6197bad871757c950cac4329b63aff8b171fdbdb3fe6",
			Size:         69458,
			ModifiedOn:   time.Now().Add(-time.Minute),
		},
	})
	text := out.String()
	assert.True(t, strings.HasPrefix(text, "uploaded\n"))
	assert.Contains(t, text, "firmware")
	assert.Contains(t, text, "stm32-app-slot-a")
	assert.Contains(t, text, "name")
	assert.Contains(t, text, "firmware.signed.bin")
	assert.Contains(t, text, "internal name")
	assert.Contains(t, text, "stm32-app-slot-a.bin")
	assert.Contains(t, text, "67.83 KiB")
	assert.Contains(t, text, "sha256:b1f3c09e4acbbf8ca40e6197bad871757c950cac4329b63aff8b171fdbdb3fe6")
	assert.Contains(t, text, "modified")
}
