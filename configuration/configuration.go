package configuration

import (
	"errors"
)

var (
	configurationShouldNotBeEmpty = errors.New("key to find configuration should not be empty")
)

type (
	config struct {
		data map[string]interface{}
	}

	Configuration interface {
		GetInt(key string) int64
		GetString(key string) string
		GetBool(key string) bool
		GetFloat(key string) float64
		GetBinary(key string) []byte
		GetArray(key string) []string
		GetMap(key string) map[string]string
	}
)
