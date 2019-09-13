package config

import (
	"os"
	"gopkg.in/yaml.v2"
)

const (
	CONFIG_FILE = "config/config.yaml"
)

type CacheConfig struct {
	RefreshInterval int `yaml:"refresh"`
}


// Config struct holds all important configuration paramters which 
// are read from config.yaml file and can be overriden by env variables
type Config struct
{
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`

	Redis struct {
		Url string `yaml:"url"`
	} `yaml:"redis"`

	UpstreamTarget struct {
	    Scheme  string `yaml:"scheme"`
	    Url     string `yaml:"url"`
	    Token   string `yaml:"token"`
	    Timeout int    `yaml:"timeout"`
	} `yaml:"target"`

	Cache CacheConfig `yaml:"cache"`

	Org struct {
		Name string `yaml: "name"`
		CachedURL []string `yaml:"cached"`
	}
}

var conf *Config

// GetServerPort returns the port the server is listening on
func (c *Config) GetServerPort() string {
	return c.Server.Port
}

func (c *Config) GetRedisURL() string {
	return c.Redis.Url
}

func (c* Config) GetTargetToken() string {
	return c.UpstreamTarget.Token
}

func (c* Config) GetTargetScheme() string {
	return c.UpstreamTarget.Scheme
}

func (c* Config) GetTargetUrl() string {
	return c.UpstreamTarget.Url
}

func (c* Config) GetTargetTimeout() int {
	return c.UpstreamTarget.Timeout
}

func (c* Config) GetCacheConfig() CacheConfig {
	return c.Cache
}

func (c *Config) GetCachedURLs() []string {
	return c.Org.CachedURL
}

func (c *Config) GetOrg() string {
	return c.Org.Name
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
	
	if os.Getenv("GITHUB_API_TOKEN") != "" {
		cfg.UpstreamTarget.Token = os.Getenv("GITHUB_API_TOKEN")
	}

	if os.Getenv("REDIS_URL") != "" {
		cfg.Redis.Url = os.Getenv("REDIS_URL")
	}
	return nil
}
