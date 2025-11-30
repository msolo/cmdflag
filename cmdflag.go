package cmdflag

import (
	"context"
	"flag"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/posener/complete"
)

var (
	PredictNothing  = complete.PredictNothing
	PredictAnything = complete.PredictAnything
	PredictDirs     = complete.PredictDirs
	PredictFiles    = complete.PredictFiles
	PredictFilesSet = complete.PredictFilesSet
	PredictSet      = complete.PredictSet
	PredictOr       = complete.PredictOr
)

type Args = complete.Args

type Command struct {
	Name      string
	UsageLine string
	UsageLong string
	Run       func(ctx context.Context, cmd *Command, args []string)
	Flags     []Flag
	Args      complete.Predictor
	flagSet   *flag.FlagSet
	bound     bool
}

type Flag struct {
	Name         string
	FlagType     int
	DefaultValue interface{}
	Usage        string
	Predictor    complete.Predictor
}

const (
	FlagTypeInt = iota
	FlagTypeString
	FlagTypeDuration
	FlagTypeBool
)

func (cmd *Command) BindFlagSet(bindFlags map[string]interface{}) *flag.FlagSet {
	if cmd.bound {
		panic("flag set already bound for command: " + cmd.Name)
	}
	fs := cmd.flagSet
	if fs == nil {
		fs = flag.NewFlagSet(cmd.Name, flag.ExitOnError)
		fs.Usage = func() {
			fmt.Fprintf(os.Stderr, "Usage of %s:\n\n%s\n", cmd.Name,
				ensureNewline(cmd.UsageLong))
			fs.PrintDefaults()
		}
	}
	for name, val := range bindFlags {
		var fdef *Flag
		for _, x := range cmd.Flags {
			if x.Name == name {
				fdef = &x
				break
			}
		}
		if fdef == nil {
			panic("attempt to bind invalid flag: " + name)
		}
		switch fdef.FlagType {
		case FlagTypeInt:
			fs.IntVar(val.(*int), fdef.Name, fdef.DefaultValue.(int), fdef.Usage)
		case FlagTypeString:
			fs.StringVar(val.(*string), fdef.Name, fdef.DefaultValue.(string), fdef.Usage)
		case FlagTypeDuration:
			fs.DurationVar(val.(*time.Duration), fdef.Name, fdef.DefaultValue.(time.Duration), fdef.Usage)
		case FlagTypeBool:
			fs.BoolVar(val.(*bool), fdef.Name, fdef.DefaultValue.(bool), fdef.Usage)
		}
	}
	cmd.flagSet = fs
	cmd.bound = len(bindFlags) > 0
	return fs
}

func (cmd *Command) FlagSet() *flag.FlagSet {
	if cmd.flagSet == nil {
		cmd.BindFlagSet(nil)
	}
	return cmd.flagSet
}

type boolFlag interface {
	IsBoolFlag() bool
}

// Follow prevailing conventions
func flagName(n string) string {
	if len(n) > 1 {
		return "--" + n
	}
	return "-" + n
}

func (cmd *Command) completeFlags() map[string]complete.Predictor {
	cf := make(map[string]complete.Predictor)
	for _, fl := range cmd.Flags {
		// bool flags are tricky. They require an = without whitespace
		// so there is no parsing ambiguity. Suggest the full option
		// rather than completing in stages, otherwise the completer
		// will inject extra whitespace at the end.
		name := flagName(fl.Name)
		if fl.FlagType == FlagTypeBool {
			cf[name+"=0"] = PredictNothing
			cf[name+"=1"] = PredictNothing
		} else {
			cf[name] = fl.Predictor
		}
	}
	fs := cmd.FlagSet()
	fs.VisitAll(func(fl *flag.Flag) {
		name := flagName(fl.Name)
		if _, ok := fl.Value.(boolFlag); ok {
			cf[name+"=0"] = PredictNothing
			cf[name+"=1"] = PredictNothing
		} else {
			// Just complete the flag name - we can't know much else.
			cf[name] = PredictNothing
		}
	})
	return cf
}

func (cmd *Command) completeCommand() complete.Command {
	return complete.Command{
		Args:  cmd.Args,
		Flags: cmd.completeFlags(),
	}
}

func ensureNewline(s string) string {
	if s == "" || s[len(s)-1] != '\n' {
		s += "\n"
	}
	return s
}

func Parse(cmdMain *Command, subCmds []*Command) (cmd *Command, args []string) {
	if flag.CommandLine.Parsed() {
		panic("flag.Parse() parse cannot be called when using cmdflag")
	}

	cmdModeMap := make(map[string]*Command)
	cmplModeMap := make(complete.Commands)
	for _, cmd := range subCmds {
		cmdModeMap[cmd.Name] = cmd
		cmplModeMap[cmd.Name] = cmd.completeCommand()
	}

	if cmdMain.flagSet != nil {
		flag.CommandLine.VisitAll(func(fl *flag.Flag) {
			cmdMain.flagSet.Var(fl.Value, fl.Name, fl.Usage)
		})
		flag.CommandLine = cmdMain.flagSet
	} else {
		cmdMain.flagSet = flag.CommandLine
	}
	flagSet := cmdMain.FlagSet()
	flagSet.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", cmdMain.Name)
		fmt.Fprintf(os.Stderr, ensureNewline(cmdMain.UsageLong))
		flagSet.PrintDefaults()
		if len(subCmds) > 0 {
			fmt.Fprintf(os.Stderr, "\nSubcommands:\n")
			tabWr := tabwriter.NewWriter(os.Stderr, 0, 4, 2, ' ', 0)
			for _, cmd := range subCmds {
				_, _ = tabWr.Write([]byte(fmt.Sprintf("\t%s:\t%s\n", cmd.Name, ensureNewline(cmd.UsageLine))))
			}
			_ = tabWr.Flush()
		}
		fmt.Fprintf(os.Stderr, "\nFor more information, use <subcommand> -help.\n")
		completionDoc := `
Install bash completions by running:
	complete -C %v %v
`
		fmt.Fprintf(os.Stderr, completionDoc, cmdMain.Name, cmdMain.Name)
	}

	cmplMain := complete.Command{
		Sub:   cmplModeMap,
		Flags: cmdMain.completeFlags(),
	}

	// Don't worry about errors here, the individual commands will reparse
	// for their specific flags.
	_ = flagSet.Parse(os.Args[1:])

	exitUsage := func() {
		flagSet.Usage()
		os.Exit(1)
	}

	args = flagSet.Args()
	cmdName := ""
	if len(args) > 0 {
		cmdName = args[0]
		args = args[1:]
	}

	hasDanglingEqual := strings.HasSuffix(os.Getenv("COMP_LINE"), "=")
	// FIXME(msolo) There is an upstream bug. When a flag complete ends in "=",
	// modes are suggested, rather than the flags. The real fix is to go
	// upstream and fix the library. Or reimplement it from scratch.
	if hasDanglingEqual {
		for _, k := range slices.Collect(maps.Keys(cmplModeMap)) {
			if k != cmdName {
				delete(cmplModeMap, k)
			}
		}
	}

	completer := complete.New(cmdMain.Name, cmplMain)
	if completer.Complete() {
		os.Exit(0)
	}

	if cmd, ok := cmdModeMap[cmdName]; ok {
		return cmd, args
	}

	if cmdName != "" {
		fmt.Fprintf(os.Stderr, "mode provided but not defined: %s\n", cmdName)
	}
	exitUsage()
	return
}
