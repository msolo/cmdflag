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
		{Name: "subflag", FlagType: cmdflag.FlagTypeBool, DefaultValue: false, Usage: "demo a subflag", Predictor: nil},
		{Name: "v", FlagType: cmdflag.FlagTypeBool, DefaultValue: false, Usage: "demo a simple subflag", Predictor: nil},
	},
}

func runDemo(ctx context.Context, cmd *cmdflag.Command, args []string) {
	subflag := false
	flags := cmd.BindFlagSet(map[string]interface{}{"subflag": &subflag})
	_ = flags.Parse(args)
	fmt.Println("demo stdflag:", *stdflag, "subflag:", subflag)
}

var cmdMain = &cmdflag.Command{
	Name:      "example",
	UsageLong: doc,
	Flags: []cmdflag.Flag{
		{Name: "timeout", FlagType: cmdflag.FlagTypeDuration, DefaultValue: 0 * time.Millisecond, Usage: "timeout for command execution", Predictor: nil},
		{Name: "config-file", FlagType: cmdflag.FlagTypeString, DefaultValue: "", Usage: "local config file", Predictor: cmdflag.PredictFiles("*")},
	},
}

var subcommands = []*cmdflag.Command{
	cmdDemo,
}

var stdflag = flag.Bool("std-flag", false, "Just a placeholder to show standard flag support.")

func main() {
	var timeout time.Duration
	var configFile string

	cmdMain.BindFlagSet(map[string]any{"timeout": &timeout,
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
