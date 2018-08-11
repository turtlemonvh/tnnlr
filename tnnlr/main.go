package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/timjchin/unpuzzled"
	"github.com/turtlemonvh/tnnlr"
)

// Config

func main() {
	logLevels := make([]string, len(logrus.AllLevels))
	for il, l := range logrus.AllLevels {
		logLevels[il] = l.String()
	}

	myTnnlr := &tnnlr.Tnnlr{}
	app := unpuzzled.NewApp()
	app.Command = &unpuzzled.Command{
		Name: "tnnlr",
		Variables: []unpuzzled.Variable{
			&unpuzzled.StringVariable{
				Name:        "log-level",
				Destination: &(myTnnlr.LogLevel),
				Description: fmt.Sprintf("Logging levels. Options are: [%s]", strings.Join(logLevels, ",")),
				Default:     "info",
			},
			&unpuzzled.StringVariable{
				Name:        "tunnels",
				Destination: &(myTnnlr.TunnelReloadFile),
				Description: "Configuration file listing tunnels. This can be read from and written to via the web UI.",
				Default:     ".tnnlr",
			},
			&unpuzzled.StringVariable{
				Name:        "ssh-exec",
				Destination: &(myTnnlr.SshExec),
				Description: "The executable process to use for ssh. Can be a full path or just a command name that works in your shell.",
				Default:     "ssh",
			},
			&unpuzzled.IntVariable{
				Name:        "port",
				Destination: &(myTnnlr.Port),
				Description: "The port to run the server on for the web UI.",
				Default:     8080,
			},
		},
		Action: func() {
			// Run web server
			myTnnlr.Init()
			myTnnlr.Run()
		},
		Subcommands: []*unpuzzled.Command{
			/*
			&unpuzzled.Command{
				Name:      "ls",
				Usage:     "List running tunnels",
				Variables: []unpuzzled.Variable{},
				Action: func() {
					myTnnlr.Init()
					// NOT IMPLEMENTED
				},
			},
			*/
		},
	}
	app.Authors = []unpuzzled.Author{
		{
			Name: "Timothy Van Heest",
			Email: "timothy@ionic.com",
		},
	}
	app.ParsingOrder = []unpuzzled.ParsingType{
		unpuzzled.EnvironmentVariables,
		unpuzzled.CliFlags,
	}
	app.Run(os.Args)
}
