package helpers

var cases = []int{2, 0, 1, 1, 1, 2}

func Declension(num int, words []string) string {
	pos := 0

	if num%100 > 4 && num%100 < 20 {
		pos = 2
	} else if num%10 < 5 {
		pos = cases[num%10]
	} else {
		pos = cases[5]
	}
	return words[pos]
}
