package conf

import "fmt"

// Config carries command line configurations
type Config struct {
	Port   int
	Target string
	IP     string
	RPM    int64
}

// Addr is a convenience function that generates
// a combination of the configured IP and port in the format of
// IP:PORT
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.IP, c.Port)
}

var defaultCfg Config = Config{
	Port: 2400,
	IP:   "127.0.0.1",
	RPM:  100,
}

// Cfg is our main configuration variable, this holds all the service's configuration
// options
var Cfg *Config = &Config{}
