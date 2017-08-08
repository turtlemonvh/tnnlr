package main

import (
	"os"

	"github.com/timjchin/unpuzzled"
	"github.com/turtlemonvh/tnnlr"
)

// Config

func main() {
	myTnnlr := &tnnlr.Tnnlr{}
	app := unpuzzled.NewApp()
	app.Command = &unpuzzled.Command{
		Name: "tnnlr",
		Variables: []unpuzzled.Variable{
			&unpuzzled.StringVariable{
				Name:        "log-level",
				Destination: &(myTnnlr.LogLevel),
				Default:     "info",
			},
		},
		Action: func() {
			// Run website
			myTnnlr.Init()
			myTnnlr.Run()
		},
		Subcommands: []*unpuzzled.Command{
			&unpuzzled.Command{
				Name:      "ls",
				Usage:     "List running tunnels",
				Variables: []unpuzzled.Variable{},
				Action: func() {
					myTnnlr.Init()
					// NOT IMPLEMENTED
				},
			},
		},
	}
	app.Run(os.Args)
}
