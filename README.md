# cmdflag
Simple wiring for command flags that enables subcommands and shell completions.

The heavy lifting of completion is accomplished using https://github.com/posener/complete. cmdflags just makes the API more straightforward for a broad class of programs.

# An Example Self-completing Program
A runnable example is included in [example.go](./example/example.go).

`example` is a simple program that generates its own completions.

```go
package main

import (
	"context"
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
	fmt.Println("demo subflag:", subflag)
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

func main() {
	var timeout time.Duration
	var configFile string

	cmdMain.BindFlagSet(map[string]interface{}{
		"timeout": &timeout,
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
```

# Enabling Command Completion in Bash
```sh
> cd example/
> go build
> ./example 
Usage of example:
example - a self-completing program
  -config-file string
    	local config file
  -timeout duration
    	timeout for command execution

Install bash completions by running:
	complete -C example example
> complete -C ./example example
> ./example -<tab>
-config-file  -timeout      
> ./example <tab>demo 
```

## Other Shells
zsh and fish are supported by the upstream completiong libraries, so this should work.
