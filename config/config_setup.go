package config

import (
	"os"
	"gopkg.in/yaml.v2"
	"github.com/kelseyhightower/envconfig"
)

const (
	CONFIG_FILE = "config/config.yaml"
)

// Config struct holds all important configuration paramters which 
// are read from config.yaml file and can be overriden by env variables
type Config struct
{
	Server struct {
		Port string `yaml:"port", envconfig:"SERVER_PORT"`
	} `yaml:"server"`
	Redis struct {
		Url string `yaml:"url", envconfig:"REDIS_URL"`
	} `yaml:"REDIS_URL"`
}

var conf *Config

// GetServerPort returns the port the server is listening on
func (c *Config) GetServerPort() string {
	return c.Server.Port
}

// InitConfig initializes the config object for our program. Typically to be called before starting server instance
func InitConfig() (*Config, error) {

	conf = &Config{}

	// reads the configuration from yaml file
	if err := readFromFile(conf); err != nil {
		return conf, err
	}
	
	// overrides the configuration set in env variables
	if err := readFromEnv(conf); err != nil {
		return conf, err
	}

	return conf, nil
}

// GetConfig provides public access to config
func GetConfig() *Config {
	return conf
}

// reads the config yaml file and sets the config struct
func readFromFile(cfg *Config) (err error) {
	
	// read the configuration yaml file
	fp, err := os.Open(CONFIG_FILE)
	if err != nil {
		return err
	}	
	
	decoder := yaml.NewDecoder(fp)
	if err := decoder.Decode(cfg); err != nil {
		return err
	}

	return nil
}

// reads from env variables and assigns it to struct
func readFromEnv(cfg *Config) (err error) {
	
	if err := envconfig.Process("", cfg); err != nil {
		return err
	}

	return nil
}
