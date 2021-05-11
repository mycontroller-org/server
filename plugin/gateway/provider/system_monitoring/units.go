package systemmonitoring

func getValueByUnit(value uint64, unit string) uint64 {
	switch unit {
	case "KiB":
		return value / 1024

	case "MiB":
		return value / (1024 * 1024)

	case "GiB":
		return value / (1024 * 1024 * 1024)

	case "TiB":
		return value / (1024 * 1024 * 1024 * 1024)

	default:
		return value
	}
}
