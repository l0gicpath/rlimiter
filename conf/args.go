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

func DisplayUsage() {
	fmt.Fprintf(os.Stderr, usageBanner, os.Args[0])
	flag.PrintDefaults()
}

func setupFlags() {
	flag.IntVar(&Cfg.Port, "port", defaultCfg.Port, portArgUsage)
	flag.StringVar(&Cfg.IP, "ip", defaultCfg.IP, ipArgUsage)
	flag.StringVar(&Cfg.Target, "target", defaultCfg.Target, targetArgUsage)
	flag.Int64Var(&Cfg.RPM, "rpm", defaultCfg.RPM, rpmArgUsage)
}

func CommandLineArgs() error {
	setupFlags()

	flag.Usage = DisplayUsage
	flag.Parse()

	if Cfg.Target == "" {
		return errors.New("Please specify a -target option")
	}

	return nil
}
