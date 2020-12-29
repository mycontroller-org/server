package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// See: http://en.wikipedia.org/wiki/Binary_prefix
const (
	// Decimal
	KB = 1000
	MB = 1000 * KB
	GB = 1000 * MB
	TB = 1000 * GB
	PB = 1000 * TB

	// Binary
	KiB = 1024
	MiB = 1024 * KiB
	GiB = 1024 * MiB
	TiB = 1024 * GiB
	PiB = 1024 * TiB
)

type unitMap map[string]int64

var (
	decimalMap = unitMap{"k": KB, "m": MB, "g": GB, "t": TB, "p": PB}
	binaryMap  = unitMap{"k": KiB, "m": MiB, "g": GiB, "t": TiB, "p": PiB}
	sizeRegex  = regexp.MustCompile(`^(\d+(\.\d+)*) ?([kKmMgGtTpP])?[iI]?[bB]?$`)
)

var decimapAbbrs = []string{"B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
var binaryAbbrs = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}

func getSizeAndUnit(size float64, base float64, _map []string) (float64, string) {
	i := 0
	unitsLimit := len(_map) - 1
	for size >= base && i < unitsLimit {
		size = size / base
		i++
	}
	return size, _map[i]
}

// CustomSize returns a human-readable approximation of a size
// using custom format.
func CustomSize(format string, size float64, base float64, _map []string) string {
	size, unit := getSizeAndUnit(size, base, _map)
	return fmt.Sprintf(format, size, unit)
}

// ToDecimalSizeStringWithPrecision allows the size to be in any precision,
func ToDecimalSizeStringWithPrecision(size float64, precision int) string {
	size, unit := getSizeAndUnit(size, 1000.0, decimapAbbrs)
	return fmt.Sprintf("%.*g%s", precision, size, unit)
}

// ToDecimalSizeString with 2 decimal
func ToDecimalSizeString(size float64) string {
	return ToDecimalSizeStringWithPrecision(size, 2)
}

// ToBinarySizeString returns a human-readable size in bytes, kibibytes,
// mebibytes, gibibytes, or tebibytes (eg. "44kiB", "17MiB").
func ToBinarySizeString(size float64) string {
	return CustomSize("%.4g%s", size, 1024.0, binaryAbbrs)
}

// ParseSizeWithDefault returns with default
func ParseSizeWithDefault(size string, defaultSize int64) int64 {
	parsed, err := ParseSize(size)
	if err != nil {
		return defaultSize
	}
	return parsed
}

// ParseSize returns an integer from a human-readable specification of a
// size using SI standard (eg. "44kB", "17MB", "17MiB").
func ParseSize(sizeStr string) (int64, error) {
	matches := sizeRegex.FindStringSubmatch(sizeStr)
	if len(matches) != 4 {
		return -1, fmt.Errorf("invalid size: '%s'", sizeStr)
	}
	fmt.Println("matches", matches)
	size, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return -1, err
	}

	unitPrefix := strings.ToLower(matches[3])

	unitMap := decimalMap
	if strings.Contains(strings.ToLower(matches[0]), "i") {
		unitMap = binaryMap
	}

	if mul, found := unitMap[unitPrefix]; found {
		size *= float64(mul)
	}

	return int64(size), nil
}
