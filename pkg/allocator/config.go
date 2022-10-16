package allocator

import (
	"errors"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Names struct {
		File string   `yaml:"file"`
		Data []string `yaml:"data"`
	}
	Passwords struct {
		File string   `yaml:"file"`
		Data []string `yaml:"data"`
	}
	CanAllocateSelf bool          `yaml:"canAllocateSelf"`
	Timeout         time.Duration `yaml:"timeout,omitempty"`
	Rules           []Rule        `yaml:"rules"`
}

type Rule struct {
	Name      string   `yaml:"name"`
	CannotGet []string `yaml:"cannotGet"`
}

var DefaultConfig = Config{
	CanAllocateSelf: false,
	Timeout:         time.Second * 5,
}

func LoadConfigFromYaml(yamlData []byte) (*Config, error) {

	c := DefaultConfig
	err := yaml.Unmarshal(yamlData, &c)
	if err != nil {
		return nil, err
	}

	err = c.validateConfig()
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Config) validateConfig() error {
	// Check we have some names
	if c.Names.File == "" && len(c.Names.Data) <= 0 {
		return errors.New("must provide either a filename or name data in config")
	}

	if c.Passwords.File == "" && len(c.Passwords.Data) <= 0 {
		return errors.New("must provide either a filename or password data in config")
	}

	if c.Timeout <= 0 {
		return errors.New("timeout must have a value greater than 0")
	}

	return nil
}
