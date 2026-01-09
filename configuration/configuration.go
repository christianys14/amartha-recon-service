package configuration

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

var (
	configurationShouldNotBeEmpty = errors.New("key to find configuration should not be empty")
)

type config struct {
	data map[string]interface{}
}

type Configuration interface {
	GetInt(key string) int64
	GetString(key string) string
	GetBool(key string) bool
	GetFloat(key string) float64
	GetBinary(key string) []byte
	GetArray(key string) []string
	GetMap(key string) map[string]string
}

func (c *config) GetInt(key string) int64 {
	value, ok := c.data[key]
	if ok {
		str := fmt.Sprintf("%s", value)
		num, err := strconv.ParseInt(str, 10, 64)
		if err == nil {
			return num
		}
	}
	return 0
}

func (c *config) GetString(key string) string {
	value, ok := c.data[key]
	if ok {
		str, ok := value.(string)
		if ok {
			return str
		}
	}
	return ""
}

func (c *config) GetBool(key string) bool {
	value, ok := c.data[key]
	if ok {
		str, ok := value.(string)
		if ok {
			boolean, err := strconv.ParseBool(str)
			if err == nil {
				return boolean
			}
		}
	}
	return false
}

func (c *config) GetFloat(key string) float64 {
	value, ok := c.data[key]
	if ok {
		str, ok := value.(string)
		if ok {
			num, err := strconv.ParseFloat(str, 64)
			if err == nil {
				return num
			}
		}
	}
	return 0
}

func (c *config) GetBinary(key string) []byte {
	value, ok := c.data[key]
	if ok {
		str, ok := value.(string)
		if ok {
			bytes, err := base64.StdEncoding.DecodeString(str)
			if err == nil {
				return bytes
			}
		}
	}
	return nil
}

func (c *config) GetArray(key string) []string {
	value, ok := c.data[key]
	if ok {
		str, ok := value.(string)
		if ok {
			if str != "" {
				return strings.Split(str, ",")
			}
		}
	}
	return nil
}

func (c *config) GetMap(key string) map[string]string {
	value, ok := c.data[key]
	if ok {
		str, ok := value.(string)
		if ok {
			maps := make(map[string]string)
			array := strings.Split(str, ",")
			for _, element := range array {
				kv := strings.SplitN(element, ":", 2)
				if len(kv) == 2 {
					maps[kv[0]] = kv[1]
				}
			}
			return maps
		}
	}
	return nil
}

func FindConfiguration(key string) (Configuration, error) {
	if "" == key {
		return nil, configurationShouldNotBeEmpty
	}

	return newConfig(fmt.Sprintf("./%s.json", key))
}

func newConfig(path string) (Configuration, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("error getting file: ", err)
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(file, &result)
	if err != nil {
		log.Println("error unmarshal data: ", err)
		return nil, err
	}

	var cfg config
	cfg.data = result

	return &cfg, nil
}
