package conf

import "fmt"

// Config carries command line configurations
type Config struct {
	Port   int
	Target string
	IP     string
	RPM    int64
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.IP, c.Port)
}

var defaultCfg Config = Config{
	Port: 2400,
	IP:   "127.0.0.1",
	RPM:  100,
}
var Cfg *Config = &Config{}
