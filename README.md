# cmdflag
Simple wiring for command flags that enables subcommands and shell completions.

A runnable example is included in [example.go](./example/example.go).

`example` is a simple program that generates its own completions.

```bash
msolo@x:…/cmdflag> cd example/
msolo@x:…/cmdflag/example> go build
msolo@x:…/cmdflag/example> ./example 
mode provided but not defined: 
Usage of example:
example - a self-completing program
  -config-file string
    	local config file
  -timeout duration
    	timeout for command execution

Install bash completions by running:
	complete -C example example
msolo@x:…/cmdflag/example> complete -C ./example example
msolo@x:…/cmdflag/example> ./example -<tab>
-config-file  -timeout      
msolo@x:…/cmdflag/example> ./example <tab>demo 
```
