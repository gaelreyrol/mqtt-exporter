package mqttexporter

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Broker string
	Topics []Topic
}

type Topic struct {
	Name     string
	Interval int16
	Fields   []string
}

func NewConfig() *Config {
	return &Config{
		Broker: ":1883",
		Topics: nil,
	}
}

func (c *Config) decode(data []byte) error {
	if _, err := toml.Decode(string(data), c); err != nil {
		return err
	}

	return nil
}

func (c *Config) FromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not read config file from %s", filePath)
	}

	if err = c.decode(data); err != nil {
		return fmt.Errorf("could not decode config file from %s", filePath)
	}

	return nil
}

func (t *Topic) IsFieldFiltered(key string) bool {
	if t.Fields == nil || len(t.Fields) == 0 {
		return false
	}

	for _, field := range t.Fields {
		if field == key {
			return false
		}
	}

	return true
}
