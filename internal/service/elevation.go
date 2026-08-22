package service

func ElevationBand(value float64) string {
	switch {
	case value < 500:
		return "lowland"
	case value < 1500:
		return "montane"
	case value < 3000:
		return "alpine"
	default:
		return "high-alpine"
	}
}

func ElevationValid(value float64) bool { return value >= -500 && value <= 9000 }
