package gettUtils

import (
	"github.com/gofrs/uuid"
)

// newUUID generates a random UUID according to RFC 4122
func NewUUID() (string, error) {
	newUUID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return newUUID.String(), nil
}
