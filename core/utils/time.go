package utils

import (
	"time"
)

// ToDatetimeMillion time format
func ToDatetimeMillion(t time.Time) string {
	s := t.Format("2006-01-02 15:04:05.000")
	return s
}

// ToDatetime time format
func ToDatetime(t time.Time) string {
	s := t.Format("2006-01-02 15:04:05")
	return s
}
