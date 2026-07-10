// Package config resolves runtime configuration with precedence
// CLI flag > environment variable > built-in default, in one place.
package config

import "os"

type Config struct {
	Host     string
	Port     string
	LogLevel string
}

const (
	EnvHost     = "MYTOOLKIT_HOST"
	EnvPort     = "MYTOOLKIT_PORT"
	EnvLogLevel = "MYTOOLKIT_LOG_LEVEL"

	DefaultHost     = "0.0.0.0"
	DefaultPort     = "8080"
	DefaultLogLevel = "info"
)

// Flags carries the CLI flag values and whether each was explicitly set.
type Flags struct {
	Host        string
	HostChanged bool
	Port        string
	PortChanged bool
	LogLevel    string
	LogChanged  bool
}

// Load resolves the final Config from flags, environment variables, and defaults.
func Load(f Flags) Config {
	return Config{
		Host:     resolve(f.Host, f.HostChanged, EnvHost, DefaultHost),
		Port:     resolve(f.Port, f.PortChanged, EnvPort, DefaultPort),
		LogLevel: resolve(f.LogLevel, f.LogChanged, EnvLogLevel, DefaultLogLevel),
	}
}

func resolve(flagValue string, flagChanged bool, envVar, def string) string {
	if flagChanged {
		return flagValue
	}
	if v, ok := os.LookupEnv(envVar); ok && v != "" {
		return v
	}
	return def
}

func (c Config) Addr() string {
	return c.Host + ":" + c.Port
}
