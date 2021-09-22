package utils

import uuid "github.com/satori/go.uuid"

// NewUuid return union id
func NewUuid() string {
	return uuid.NewV4().String()
}

// ParseUuid check uuid valid
func ParseUuid(s string) bool {
	_, err := uuid.FromString(s)
	if err != nil {
		return false
	}
	return true
}
