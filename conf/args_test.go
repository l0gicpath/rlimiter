package conf

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestAddr(t *testing.T) {
	config := &Config{Port: 2222, IP: "1.1.1.1"}
	assert.Equal(t, config.Addr(), "1.1.1.1:2222")
}

func emptyConfig() bool {
	return Cfg.Port == 0 &&
		Cfg.IP == "" &&
		Cfg.Target == "" &&
		Cfg.RPM == 0
}

func resetCommandLine() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	CommandLineArgs()
}

func TestCommandArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	assert.Equal(t, emptyConfig(), true)

	os.Args = []string{"test", "-port=9091"}
	resetCommandLine()
	assert.Equal(t, Cfg.Port, 9091)

	os.Args = []string{"test", "-ip=1.1.1.2"}
	resetCommandLine()
	assert.Equal(t, Cfg.IP, "1.1.1.2")

	os.Args = []string{"test", "-target=1.1.1.1:90"}
	resetCommandLine()
	assert.Equal(t, Cfg.Target, "1.1.1.1:90")

	os.Args = []string{"test", "-rpm=5"}
	resetCommandLine()
	assert.Equal(t, Cfg.RPM, int64(5))
}

func TestRequiredTargetArg(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"test", "-ip=1.1.1.1", "-port=90"}
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	err := CommandLineArgs()
	assert.NotNil(t, err)
}
