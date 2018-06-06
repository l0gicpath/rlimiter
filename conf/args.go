package conf

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

const (
	portArgUsage   = "port number to bind to"
	ipArgUsage     = "ip address to bind to"
	targetArgUsage = "address of server we are protecting"
	rpmArgUsage    = "rate of connections allowed per minute"
	usageBanner    = "Usage: %s -target=TARGET [-option=value...]\nOptions:\n"
)

// DisplayUsage is our own function for displaying help information
// on the CLI interface
func DisplayUsage() {
	fmt.Fprintf(os.Stderr, usageBanner, os.Args[0])
	flag.PrintDefaults()
}

// setupFlags will set the command line interface flags
// originally splitted away from CommandLineArgs for testing
// but now I see no reason for that. :TODO: Refactor
func setupFlags() {
	flag.IntVar(&Cfg.Port, "port", defaultCfg.Port, portArgUsage)
	flag.StringVar(&Cfg.IP, "ip", defaultCfg.IP, ipArgUsage)
	flag.StringVar(&Cfg.Target, "target", defaultCfg.Target, targetArgUsage)
	flag.Int64Var(&Cfg.RPM, "rpm", defaultCfg.RPM, rpmArgUsage)
}

// CommandLineArgs sets up the command line interface, setting up the flags
// and parses them. It will return an error if -target option is missing
// since it's required.
func CommandLineArgs() error {
	setupFlags()

	flag.Usage = DisplayUsage
	flag.Parse()

	if Cfg.Target == "" {
		return errors.New("Please specify a -target option")
	}

	return nil
}
