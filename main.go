package main

import (
	"os"

	org "github.com/bash-sh/photos/organize"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	cli "github.com/urfave/cli/v2"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "print only the version",
	}
	app := &cli.App{
		Name:    "photos",
		Version: "0.2.1",
		Usage:   "organize a photo library",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "toggle debug logging",
			},
		},
		Action: func(c *cli.Context) error {
			log.Fatal().Msg("Please select a command or refer to the help")
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "organize",
				Aliases: []string{"org"},
				Usage:   "organizes a library",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "InPath",
						Aliases: []string{"in"},
						Usage:   "source photos from `PATH`",
					},
					&cli.StringFlag{
						Name:    "OutPath",
						Aliases: []string{"out"},
						Usage:   "export photos to `PATH`",
					},
					&cli.StringFlag{
						Name:    "Topic",
						Aliases: []string{"top"},
						Usage:   "`TOPIC` of the processed photos (e.g., location, event)",
					},
				},
				Action: func(c *cli.Context) error {
					if c.Bool("debug") {
						zerolog.SetGlobalLevel(zerolog.DebugLevel)
					}
					lib := new(org.Library)
					if c.NumFlags() == 6 {
						lib.InPath = c.String("InPath")
						lib.OutPath = c.String("OutPath")
						lib.Topic = c.String("Topic")
					} else {
						lib.Init()
					}
					lib.Validate()
					lib.Process()
					return nil
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err).Msg("App cannot be run")
	}
}
