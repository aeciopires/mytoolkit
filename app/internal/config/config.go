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

const (
	EnvMCPTransport = "MYTOOLKIT_MCP_TRANSPORT"
	EnvMCPPort      = "MYTOOLKIT_MCP_PORT"

	DefaultMCPTransport = "stdio"
	DefaultMCPPort      = "8081"
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

// MCPFlags carries the CLI flag values for the `mcp` subcommand and
// whether each was explicitly set. Host and LogLevel reuse the same
// env vars/defaults as Flags (EnvHost/EnvLogLevel) since they mean the
// same thing for either server; only Transport and Port are MCP-specific.
type MCPFlags struct {
	Transport        string
	TransportChanged bool
	Host             string
	HostChanged      bool
	Port             string
	PortChanged      bool
	LogLevel         string
	LogChanged       bool
}

// MCPConfig is the resolved configuration for `mytoolkit mcp`.
type MCPConfig struct {
	Transport string
	Host      string
	Port      string
	LogLevel  string
}

// LoadMCP resolves the final MCPConfig from flags, environment variables,
// and defaults, following the same flag > env > default precedence as Load.
func LoadMCP(f MCPFlags) MCPConfig {
	return MCPConfig{
		Transport: resolve(f.Transport, f.TransportChanged, EnvMCPTransport, DefaultMCPTransport),
		Host:      resolve(f.Host, f.HostChanged, EnvHost, DefaultHost),
		Port:      resolve(f.Port, f.PortChanged, EnvMCPPort, DefaultMCPPort),
		LogLevel:  resolve(f.LogLevel, f.LogChanged, EnvLogLevel, DefaultLogLevel),
	}
}

func (c MCPConfig) Addr() string {
	return c.Host + ":" + c.Port
}
