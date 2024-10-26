package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"sync"

	"github.com/urfave/cli/v2"
	"golang.org/x/sync/semaphore"
)

func cDirTreeScan(root string, concurrency uint64) error {
	var (
		poolSize       uint64 = concurrency
		chSize         uint64 = poolSize * 2
		rwg            sync.WaitGroup
		swg            sync.WaitGroup
		dirs           []string
		files          []string
		errs           []error
		drch                               = make(chan string, chSize)
		frch                               = make(chan string, chSize)
		ech                                = make(chan error, chSize)
		poolSem        *semaphore.Weighted = semaphore.NewWeighted(int64(poolSize))
		ctx            context.Context     = context.TODO()
		scanDirTree    func(string)
		collectResults func()
	)

	scanDirTree = func(dir string) {
		defer swg.Done()

		_ = poolSem.Acquire(ctx, 1)
		defer poolSem.Release(1)

		drch <- dir
		entries, err := os.ReadDir(dir)
		if err != nil {
			ech <- err
		}
		for _, entry := range entries {
			path := path.Join(dir, entry.Name())
			if entry.IsDir() {
				swg.Add(1)
				go scanDirTree(path)
			} else {
				frch <- path
			}
		}
	}

	collectResults = func() {
		defer rwg.Done()

		rwg.Add(1)
		go func() {
			defer rwg.Done()
			for err := range ech {
				errs = append(errs, err)
			}
		}()

		rwg.Add(1)
		go func() {
			defer rwg.Done()
			for dir := range drch {
				fmt.Printf("d %v\n", dir)
				dirs = append(dirs, dir)
			}
		}()

		rwg.Add(1)
		go func() {
			defer rwg.Done()
			for file := range frch {
				fmt.Printf("f %v\n", file)
				files = append(files, file)
			}
		}()
	}

	rwg.Add(1)
	go collectResults()

	swg.Add(1)
	go scanDirTree(root)

	// Wait for all scans to be ready
	swg.Wait()
	close(drch)
	close(frch)
	close(ech)

	// Wait for the collector to be ready
	rwg.Wait()

	if len(errs) > 0 {
		return errors.New("one or more errors happened during scanning")
	}
	for _, f := range errs {
		fmt.Println(f)
	}
	return nil
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
					if v <= 1 {
						return fmt.Errorf("concurrency value %v is not greater than or equal to 1", v)
					}
					return nil
				},
			},
		},
		Action: func(cCtx *cli.Context) error {
			var root string

			if cCtx.NArg() > 0 {
				root = cCtx.Args().Get(0)
			} else {
				return cli.Exit("no directory to scan", 1)
			}

			fileInfo, err := os.Stat(root)
			if err != nil {
				return cli.Exit(err, 1)
			}
			if !fileInfo.IsDir() {
				return cli.Exit("the provided path is not a directory", 1)
			}

			concurrency := cCtx.Uint64("concurrency")
			if err := cDirTreeScan(root, concurrency); err != nil {
				return cli.Exit(err, 1)
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
