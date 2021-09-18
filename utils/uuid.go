package utils

import uuid "github.com/satori/go.uuid"

func NewUuid() string {
	return uuid.NewV4().String()
}

func ParseUuid(s string) bool {
	_, err := uuid.FromString(s)
	if err != nil {
		return false
	}
	return true
}