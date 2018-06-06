package main

import (
	"fmt"
	"github.com/l0gicpath/rlimiter/conf"
	"github.com/l0gicpath/rlimiter/proxy"
	"log"
	"os"
)

func main() {
	err := conf.CommandLineArgs()
	if err != nil {
		fmt.Println(err)
		conf.DisplayUsage()
		os.Exit(1)
	}

	server := proxy.New(conf.Cfg.Target, conf.Cfg.Addr())

	log.Printf("Starting rlimiter... binding to: %s", conf.Cfg.Addr())
	log.Printf("Limiting access to %s by %v reqs/m", conf.Cfg.Target, conf.Cfg.RPM)
	log.Fatal(server.HTTP.ListenAndServe())
}
