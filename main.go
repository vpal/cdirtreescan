package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/urfave/cli/v2"
	"github.com/vpal/cdirtreescan/scan"
)

func validate(cCtx *cli.Context) (root string, concurrency uint64, err error) {
	if cCtx.NArg() != 1 {
		return root, concurrency, cli.Exit("provide exactly one directory to scan", 1)
	}
	root = cCtx.Args().Get(0)

	fileInfo, err := os.Stat(root)
	if err != nil {
		return root, concurrency, cli.Exit(err, 1)
	}

	if !fileInfo.IsDir() {
		return root, concurrency, cli.Exit("the provided path is not a directory", 1)
	}

	return root, cCtx.Uint64("concurrency"), err
}

func main() {
	app := &cli.App{
		Name:  "cdirtreescan",
		Usage: "Scan all entries within a directory",
		Flags: []cli.Flag{
			&cli.Uint64Flag{
				Name:    "concurrency",
				Value:   uint64(runtime.NumCPU() * 2),
				Aliases: []string{"c"},
				Usage:   "upper limit of the number of concurrent scans",
				Action: func(ctx *cli.Context, v uint64) error {
					if v < 1 {
						return fmt.Errorf("concurrency value %v is not greater than or equal to 1", v)
					}
					return nil
				},
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "count",
				Aliases: []string{"cnt"},
				Usage:   "Count the number of directories and files",
				Action: func(cCtx *cli.Context) error {
					root, concurrency, err := validate(cCtx)
					if err != nil {
						return err
					}
					ds := scan.NewDirTreeScanner(cCtx.Context, root, concurrency)
					ds.Count()
					return nil
				},
			},
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "List directories and files",
				Action: func(cCtx *cli.Context) error {
					root, concurrency, err := validate(cCtx)
					if err != nil {
						return err
					}
					ds := scan.NewDirTreeScanner(cCtx.Context, root, concurrency)
					ds.List()
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
