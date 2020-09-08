package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/msolo/cmdflag"
)

const (
	doc = `example - a self-completing program`
)

var cmdDemo = &cmdflag.Command{
	Name:      "demo",
	Run:       runDemo,
	UsageLine: "example demo",
	UsageLong: `Run a simple demo.`,
	Flags: []cmdflag.Flag{
		{"subflag", cmdflag.FlagTypeBool, false, "demo a subflag", nil},
	},
}

func runDemo(ctx context.Context, cmd *cmdflag.Command, args []string) {
	subflag := false
	flags := cmd.BindFlagSet(map[string]interface{}{"subflag": &subflag})
	flags.Parse(args)
	fmt.Println("demo stdflag:", *stdflag, "subflag:", subflag)
}

var cmdMain = &cmdflag.Command{
	Name:      "example",
	UsageLong: doc,
	Flags: []cmdflag.Flag{
		{"timeout", cmdflag.FlagTypeDuration, 0 * time.Millisecond, "timeout for command execution", nil},
		{"config-file", cmdflag.FlagTypeString, "", "local config file", cmdflag.PredictFiles("*")},
	},
}

var subcommands = []*cmdflag.Command{
	cmdDemo,
}

var stdflag = flag.Bool("std-flag", false, "Just a placeholder to show standard flag support.")

func main() {
	var timeout time.Duration
	var configFile string

	cmdMain.BindFlagSet(map[string]interface{}{"timeout": &timeout,
		"config-file": &configFile})

	cmd, args := cmdflag.Parse(cmdMain, subcommands)
	ctx := context.Background()
	if timeout > 0 {
		nctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		ctx = nctx
	}

	cmd.Run(ctx, cmd, args)
}
