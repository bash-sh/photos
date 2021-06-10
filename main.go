package main

import (
	"fmt"
	"log"
	"os"

	org "github.com/bash-sh/photos/organize"
	cli "github.com/urfave/cli/v2"
)

func main() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "print only the version",
	}
	app := &cli.App{
		Name:    "photos",
		Version: "0.1.0",
		Usage:   "organize a photo library",
		Action: func(c *cli.Context) error {
			fmt.Println("Please select a command or refer to the help")
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "organize",
				Aliases: []string{"o"},
				Usage:   "organizes a library",
				Action: func(c *cli.Context) error {
					lib := new(org.Library)
					lib.Init()
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
