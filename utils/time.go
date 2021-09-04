package utils

import (
	"time"
)

func ToDatetimeMillion(t time.Time) string {
	s := t.Format("2006-01-02 15:04:05.000")
	return s
}

func ToDatetime(t time.Time) string {
	s := t.Format("2006-01-02 15:04:05")
	return s
}

func Timing(t1 time.Time, t2 time.Time) string {
	return t2.Sub(t1).String()
}
