package command

import (
	"github.com/jessevdk/go-flags"
)

func Parse(appName string, opts interface{}, arguments []string) ([]string, error) {
	f := flags.NewParser(opts, flags.Default)
	f.Command.LongDescription = ""
	f.Command.Name = appName
	return f.ParseArgs(arguments)
}

func ParseWithDescription(appName string, opts interface{}, arguments []string, description string) ([]string, error) {
	f := flags.NewParser(opts, flags.Default)
	f.Command.LongDescription = description
	f.Command.Name = appName
	return f.ParseArgs(arguments)
}
