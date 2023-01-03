package main

import (
	"fmt"
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
                Aliases:     []string{"o", "out"},
				Usage:       "Dir to download into",
				DefaultText: "current directory",
                Value:       ".",
			},
			&cli.IntFlag{
				Name:        "depth",
				Aliases:     []string{"d"},
				Usage:       "Number of recursions for subfolders",
				DefaultText: "all subfolders",
                Value: -1,
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

            token, err := ReadGitToken()
            if err == nil && token != "" {
                info.ApiToken = token
            }

            results := utils.Download(info, cCtx.Path("outputDir"), cCtx.Int("depth"))

			// awaiting results from channel and loging them
			for result := range results {
				if result.Err != nil {
					if result.Context != nil {
						log.Fatalln("Error downloading " + result.Context.String())
					}
					return result.Err
				}
				if result.Context != nil && cCtx.Bool("verbose") {
					println("Downloaded " + result.Context.String())
				}
			}
			return nil
		},
        Commands: []*cli.Command{
            {
                Name: "token",
                Usage: "manage api tokens",
                Action: func(cCtx *cli.Context) error {
                    token, err := ReadGitToken()
                    if err != nil {
                        return err
                    }

                    if token == "" {
                        fmt.Println("No token set")
                    }

                    fmt.Println(token)
                    return nil
                },
                Subcommands: []*cli.Command{
                    {
                        Name: "add",
                        Usage: "add github api token",
                        Action: func(cCtx *cli.Context) error {
                            return WriteGitToken(cCtx.Args().First()) 
                        },
                    },
                    {
                        Name: "remove",
                        Usage: "remove github api token",
                        Action: func(cCtx *cli.Context) error {
                            return WriteGitToken("")
                        },
                    },
                },
            },
        },
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
