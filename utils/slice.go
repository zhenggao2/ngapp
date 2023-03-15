package utils

import "math"

// ContainsInt checks whether 't' exist in a slice of int
func ContainsInt(s []int, t int) bool {
	for _, v := range s {
		if v == t {
			return true
		}
	}

	return false
}

// IndexInt returns the index of 't' in a slice of int, and -1 if 't' is not contained in the slice
func IndexInt(s []int, t int) int {
	for i, v := range s {
		if v == t {
			return i
		}
	}

	return -1
}

// MaxInt returns the maximum of a slice of int
func MaxInt(s []int) int {
	if len(s) == 0 {
		return -1
	}

	max := s[0]
	for _, v := range s {
		if v > max {
			max = v
		}
	}

	return max
}

// MaxInt2 returns the maximum and its index of a slice of int
func MaxInt2(s []int) (int, int) {
	if len(s) == 0 {
		return -1, -1
	}

	max := s[0]
	maxi := 0
	for i, v := range s {
		if v > max {
			max = v
			maxi = i
		}
	}

	return max, maxi
}

// MinInt returns the minimum of a slice of int
func MinInt(s []int) int {
	if len(s) == 0 {
		return -1
	}

	min := s[0]
	for _, v := range s {
		if v < min {
			min = v
		}
	}

	return min
}

// MinInt2 returns the minimum and its index of a slice of int
func MinInt2(s []int) (int, int) {
	if len(s) == 0 {
		return -1, -1
	}

	min := s[0]
	mini := 0
	for i, v := range s {
		if v < min {
			min = v
			mini = i
		}
	}

	return min, mini
}

// ContainsStr checks whether 't' exists in a slice of string
func ContainsStr(s []string, t string) bool {
	for _, v := range s {
		if v == t {
			return true
		}
	}

	return false
}

// IndexStr returns the index of 't' in a slice of string, and -1 if 't' is not contained in the slice
func IndexStr(s []string, t string) int {
	for i, v := range s {
		if v == t {
			return i
		}
	}

	return -1
}

// ContainsBool checks whether 't' exist in a slice if bool
func ContainsBool(s []bool, t bool) bool {
	for _, v := range s {
		if v == t {
			return true
		}
	}

	return false
}

// IndexBool returns the index of 't' in a slice of bool, and -1 if 't' is not contained in the slice
func IndexBool(s []bool, t bool) int {
	for i, v := range s {
		if v == t {
			return i
		}
	}

	return -1
}

// FloorInt wraps math.Floor and returns an int
func FloorInt(x float64) int {
	return int(math.Floor(x))
}

// CeilInt wraps math.Ceil and returns an int
func CeilInt(x float64) int {
	return int(math.Ceil(x))
}

// RoundInt returns a rounding int
func RoundInt(x float64) int {
	return int(math.Floor(x + 0.5))
}

// SumInt returns sum of a int slice
func SumInt(s []int) int {
	sum := 0
	for _, v := range s {
		sum += v
	}

	return sum
}

// NearestInt returns the largest number that less than t and the smallest number that greater than t.
func NearestInt(s []int, t int) (int, int) {
	var lt, gt int
	for _, v := range s {
		if v > t {
			gt = v
			break
		}
	}

	for i := len(s) - 1; i >= 0; i-- {
		if s[i] < t {
			lt = s[i]
			break
		}
	}

	return lt, gt
}
