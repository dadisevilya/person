package rides

import uuid "github.com/google/uuid"

type Generator interface {
	Create() *Ride
}

type UUIDGenerator func() uuid.UUID

type service struct {
	uuidGenerator UUIDGenerator
}

type Ride struct {
	ID          string `json:"id"`
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
}

func NewService(UUIDGenerator UUIDGenerator) Generator {
	return &service{
		uuidGenerator: UUIDGenerator,
	}
}

func (s service) Create() *Ride {
	return &Ride{
		ID:          s.uuidGenerator().String(),
		Origin:      "HaBarzel Street 19, Tel Aviv-Yafo",
		Destination: "Ben Gurion International Airport",
	}
}
