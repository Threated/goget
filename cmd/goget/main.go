package main

import (
	"log"
	"os"

	"github.com/Threated/goget/pkg/utils"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:           "goget",
		Usage:          "Download a git subfolder or file",
		DefaultCommand: "help",
		Flags: []cli.Flag{
			&cli.PathFlag{
				TakesFile:   false,
				Name:        "outputDir",
				Aliases:     []string{"o"},
				Value:       ".",
				Usage:       "Dir to download into",
				DefaultText: ".",
			},
			&cli.IntFlag{
				Name:        "depth",
				Aliases:     []string{"d"},
				Usage:       "Number of recursions to download subfolders",
				DefaultText: "all subfolders",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Print names of downloaded files",
			},
		},
		Action: func(cCtx *cli.Context) error {
			info, err := utils.NewRepoInfoFromUrl(cCtx.Args().First())
			if err != nil {
				return err
			}
			results := make(chan utils.Result)
			go utils.Download(info, results)

			// awaiting results from channel and loging them
			for result := range results {
				if result.Err != nil {
					if result.Context != nil {
						println("Error downloading " + result.Context.String())
					}
					return result.Err
				}
				if result.Context != nil && cCtx.Bool("verbose") {
					println("Downloaded " + result.Context.String())
				}
			}
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
