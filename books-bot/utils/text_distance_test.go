package utils

import "testing"

func TestCloseEnoughHamming(t *testing.T) {
	cases := []struct {
		strs     []string
		expected bool
	}{
		{[]string{"hemming", "hamming"}, true},
		{[]string{"hemming", "not famming"}, false},
	}

	for _, c := range cases {
		fact := CloseEnoughHamming(c.strs[0], c.strs[1], -1)
		if fact != c.expected {
			t.Errorf("Test failed for %v pair, expected: %v, got: %v", c.strs, c.expected, fact)
		}
	}
}
