package utils

import "strconv"

// IsNum string is num true or false
func IsNum(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
