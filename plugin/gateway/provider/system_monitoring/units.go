package systemmonitoring

const (
	KiB = 1 << (10 * (iota + 1)) // Kibibytes
	MiB = 1 << (10 * (iota + 1)) // Mibibytes
	GiB = 1 << (10 * (iota + 1)) // Gibibytes
	TiB = 1 << (10 * (iota + 1)) // Tebibytes
	PiB = 1 << (10 * (iota + 1)) // Pebibytes
	EiB = 1 << (10 * (iota + 1)) // Exbibytes
)

func getValueByUnit(value uint64, unit string) float64 {
	floatValue := float64(value)

	switch unit {
	case "KiB":
		return floatValue / KiB

	case "MiB":
		return floatValue / MiB

	case "GiB":
		return floatValue / GiB

	case "TiB":
		return floatValue / TiB

	case "PiB":
		return floatValue / PiB

	case "EiB":
		return floatValue / EiB

	default:
		return floatValue
	}
}
