package utils

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

// MaxInt returns the maximum and its index of a slice of int
func MaxInt(s []int) (int, int) {
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

// MinInt returns the minimum and its index of a slice of int
func MinInt(s []int) (int, int) {
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
