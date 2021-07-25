package utils

// PyRange returns a slice similar to python's range.
func PyRange(a, b, step int) []int {
	r := make([]int, 0)
	for i := a; i < b; i += step {
		r = append(r, i)
	}

	return r
}
