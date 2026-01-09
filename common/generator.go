package common

import (
	"time"

	"github.com/google/uuid"
)

type Generate interface {
	UUID() string

	Time() time.Time
}

type generate struct {
}

//goland:noinspection GoExportedFuncWithUnexportedType
func NewGenerate() *generate {
	return &generate{}
}

func (g generate) UUID() string {
	return New().String()
}

func (g generate) Time() time.Time {
	return time.Now()
}

type UUID struct {
	Uuid uuid.UUID
}

// New generates new uuidqris
func New() UUID {
	return UUID{
		Uuid: uuid.New(),
	}
}

// String return new uuidqris of string
func (u UUID) String() string {
	return u.Uuid.String()
}
