package utils

import "math"

func NormalizedHammingDistance(str1 string, str2 string) float64 {
	n := math.Min(float64(len(str1)), float64(len(str2)))

	return float64(HammingDistance(str1, str2)) / n
}

func CloseEnoughHamming(str1 string, str2 string, threshold float64) bool {
	if threshold == -1 {
		threshold = 0.4
	}

	return NormalizedHammingDistance(str1, str2) < threshold
}

func HammingDistance(str1 string, str2 string) int {
	n := int(math.Min(float64(len(str1)), float64(len(str2))))

	distance := 0
	for i := 0; i < n; i++ {
		if str1[i] != str2[i] {
			distance++
		}
	}

	return distance
}
