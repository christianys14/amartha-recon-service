package common

import (
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/stretchr/testify/assert"
)

func Test_generate_Uuid(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "given uuid, when generate uuid, then return uuid",
			want: "04010101-0101-0101-0101-010101010102",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				g := NewGenerate()

				uuidPatch := gomonkey.ApplyFunc(
					New, func() UUID {
						return UUID{Uuid: [16]byte{04, 01, 01, 01, 01, 01, 01, 01, 01, 01, 01, 01, 01, 01, 01, 02}}
					})
				defer uuidPatch.Reset()

				assert.Equal(t, tt.want, g.UUID())
			})
	}
}

func Test_generate_Time(t *testing.T) {
	tests := []struct {
		name string
		want time.Time
	}{
		{
			name: "given time, when generate time, then return time",
			want: time.Date(2021, time.December, 20, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				g := NewGenerate()

				timenowPatch := gomonkey.ApplyFunc(
					time.Now, func() time.Time {
						return time.Date(2021, time.December, 20, 0, 0, 0, 0, time.UTC)
					})
				defer timenowPatch.Reset()

				assert.Equal(t, tt.want, g.Time())
			})
	}
}
