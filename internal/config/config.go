package config

import (
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env            string `yaml:"env" default:"development"`
	PostgresConfig `yaml:"database"`
	ServerConfig   `yaml:"server"`
}

// server config
type ServerConfig struct {
	Port        int           `yaml:"port" env:"SERVER_PORT" env-default:"8082"`
	Mode        string        `yaml:"mode" env:"SERVER_MODE" env-default:"debug"`
	Host        string        `yaml:"host" env:"SERVER_HOST" env-default:"localhost"`
	Timeout     time.Duration `yaml:"timeout" env:"SERVER_TIMEOUT" env-default:"15"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"SERVER_IDLE_TIMEOUT" env-default:"60"`
}

// postgres config
type PostgresConfig struct {
	Host     string `yaml:"host" default:"localhost"`
	Port     int    `yaml:"port" default:"5432"`
	Username string `yaml:"username" default:"postgres"`
	Password string `yaml:"password" default:"postgres"`
	Name     string `yaml:"name" default:"myappdb"`
}

// DSN constructs the Data Source Name for connecting to the PostgreSQL database
func (cfg *PostgresConfig) DSN() string {
	return "postgres://" +
		cfg.Username + ":" +
		cfg.Password + "@" +
		cfg.Host + ":" +
		strconv.Itoa(cfg.Port) + "/" +
		cfg.Name + "?sslmode=disable"
}

// ---------------------------------------------------------------------
// Get Config Path from Flag or Env
var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "Path to the config file")
}

// fetchConfigPath retrieves the configuration file path from command-line flags or environment variables
func fetchConfigPath() string {
	var res string

	if !flag.Parsed() {
		flag.Parse()
	}

	res = configPath

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	if res == "" {
		panic("config path is not provided")
	}

	return res
}

// LoadConfig loads the configuration from the specified path
func LoadConfig() Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}
	return LoadConfigFromPath(path)
}

// LoadConfigFromPath loads the configuration from a given file path
func LoadConfigFromPath(path string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic(err)
	}
	return cfg
}
