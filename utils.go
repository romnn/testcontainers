package testcontainers

import "github.com/google/uuid"

func genUUID() string {
	for {
		if id, err := uuid.NewRandom(); err == nil {
			return id.String()
		}
	}
}
