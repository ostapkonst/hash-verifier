package checksum

type FormatType int

const (
	FormatHashFirst FormatType = iota
	FormatPathFirst
)

func formatFromAlgorithm(algo Algorithm) FormatType {
	switch algo {
	case CRC32:
		return FormatPathFirst
	default:
		return FormatHashFirst
	}
}
