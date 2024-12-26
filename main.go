package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/urfave/cli/v2"
	"github.com/vpal/cdirtreescan/output"
	"github.com/vpal/cdirtreescan/scan"
)

func validateArgs(cCtx *cli.Context) error {
	if cCtx.NArg() != 1 {
		return cli.Exit("provide exactly one directory to scan", 1)
	}

	root := filepath.Clean(cCtx.Args().Get(0))
	fileInfo, err := os.Stat(root)
	if err != nil {
		return cli.Exit(err, 1)
	}

	if !fileInfo.IsDir() {
		return cli.Exit("the provided path is not a directory", 1)
	}
	cCtx.App.Metadata["root"] = root
	return nil
}

func main() {
	app := &cli.App{
		Name:  "cdirtreescan",
		Usage: "Scan all entries within a directory",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "concurrency",
				Value:   runtime.NumCPU(),
				Aliases: []string{"c"},
				Usage:   "Upper limit of the number of concurrent scans",
				Action: func(ctx *cli.Context, v int) error {
					if v < 1 {
						return fmt.Errorf("concurrency value %v is not greater than or equal to 1", v)
					}
					return nil
				},
			},
			&cli.BoolFlag{
				Name:    "suppress-errors",
				Value:   false,
				Aliases: []string{"s"},
				Usage:   "Don't display errors during scan",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "count",
				Aliases: []string{"cnt"},
				Usage:   "Count the number of directories and files",
				Before:  validateArgs,
				Action: func(cCtx *cli.Context) error {
					root := cCtx.App.Metadata["root"].(string)
					concurrency := cCtx.Int("concurrency")
					displayErrors := !cCtx.Bool("suppress-errors")

					dts, err := scan.NewDirTreeScanner(cCtx.Context, root, concurrency)
					if err != nil {
						return err
					}

					dtp := output.NewDirTreePrinter(dts, os.Stdout, os.Stderr, displayErrors)
					err = dtp.PrintCount()
					if displayErrors {
						return err
					}
					return cli.Exit("", 1)
				},
			},
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "List directories and files",
				Before:  validateArgs,
				Action: func(cCtx *cli.Context) error {
					root := cCtx.App.Metadata["root"].(string)
					concurrency := cCtx.Int("concurrency")
					displayErrors := !cCtx.Bool("suppress-errors")

					dts, err := scan.NewDirTreeScanner(cCtx.Context, root, concurrency)
					if err != nil {
						return err
					}

					dtp := output.NewDirTreePrinter(dts, os.Stdout, os.Stderr, displayErrors)
					err = dtp.PrintList()
					if displayErrors {
						return err
					}
					return cli.Exit("", 1)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
