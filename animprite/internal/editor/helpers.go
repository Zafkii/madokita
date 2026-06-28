package editor

import "strconv"

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func mathRound(v float64) float64 {
	if v < 0 {
		return float64(int(v - 0.5))
	}
	return float64(int(v + 0.5))
}

func itoa(v int) string {
	return strconv.Itoa(v)
}

func parseFloat(s string) (float64, error) {
	if s == "" || s == "-" || s == "." {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
