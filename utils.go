package testcontainers

import "github.com/google/uuid"

// UniqueID generates a unique uuid
func UniqueID() string {
	for {
		if id, err := uuid.NewRandom(); err == nil {
			return id.String()
		}
	}
}
