package allocator

import (
	"errors"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the configuration for an allocator
type Config struct {
	Names struct {
		File string   `yaml:"file" mapstructure:"file"`
		Data []string `yaml:"data" mapstructure:"data"`
	} `yaml:"names" mapstructure:"names"`
	Passwords struct {
		File string   `yaml:"file" mapstructure:"file"`
		Data []string `yaml:"data" mapstructure:"data"`
	} `yaml:"passwords" mapstructure:"passwords"`
	CanAllocateSelf bool          `yaml:"canAllocateSelf" mapstructure:"canAllocateSelf"`
	Timeout         time.Duration `yaml:"timeout,omitempty" mapstructure:"timeout"`
	Rules           []Rule        `yaml:"rules" mapstructure:"rules"`
	Name            string        `yaml:"allocation_name" mapstructure:"allocation_name"`
}

// Rule holds information about constraints for an individual name
type Rule struct {
	Name      string   `yaml:"name" mapstructure:"name"`
	CannotGet []string `yaml:"cannotGet" mapstructure:"cannotGet"`
	MustGet   string   `yaml:"mustGet" mapstructure:"mustGet"`
	Inverse   bool     `yaml:"inverse" mapstructure:"inverse"`
}

var DefaultConfig = Config{
	CanAllocateSelf: false,
	Timeout:         time.Second * 5,
}

// LoadConfigFromYaml receives yamlData and unmarshals it
// into a Config
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

// validateConfig... validates the config.
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
