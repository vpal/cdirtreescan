package scan

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"sync"
)

type DirTreeScanner struct {
	ctx         context.Context
	root        PathEntry
	chSize      uint64
	concurrency uint64
}

type PathEntry struct {
	Path  string
	Entry os.DirEntry
}

func NewDirTreeScanner(ctx context.Context, root string, concurrency uint64) (*DirTreeScanner, error) {
	fileInfo, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	return &DirTreeScanner{
		ctx: ctx,
		root: PathEntry{
			Path:  root,
			Entry: fs.FileInfoToDirEntry(fileInfo),
		},
		concurrency: concurrency,
		chSize:      concurrency * 1000,
	}, nil
}

func (dts *DirTreeScanner) Stream() (<-chan []PathEntry, <-chan error) {
	var (
		entryCh = make(chan []PathEntry, dts.chSize)
		errCh   = make(chan error, dts.chSize)
	)
	go dts.scanDirTree(entryCh, errCh)
	return entryCh, errCh
}

func (dts *DirTreeScanner) ChSize() int64 {
	return int64(dts.chSize)
}

func (dts *DirTreeScanner) scanDirTree(entryCh chan<- []PathEntry, errCh chan<- error) {
	wg := sync.WaitGroup{}
	semCh := make(chan int, dts.concurrency)
	batchSize := 256

	var scanDir func(PathEntry)
	scanDir = func(dir PathEntry) {
		defer wg.Done()

		semCh <- 1
		defer func() { <-semCh }()

		fileEntries := make([]PathEntry, 0, batchSize)

		entryCh <- []PathEntry{
			{
				Path:  dir.Path,
				Entry: dir.Entry,
			},
		}

		file, err := os.Open(dir.Path)
		if err != nil {
			errCh <- err
		}
		defer file.Close()

	Loop:
		for {
			entries, err := file.ReadDir(batchSize)
			if errors.Is(err, io.EOF) {
				break Loop
			} else if err != nil {
				errCh <- err
			}

			for _, entry := range entries {
				path := path.Join(dir.Path, entry.Name())
				if entry.IsDir() {
					wg.Add(1)
					go scanDir(PathEntry{
						Path:  path,
						Entry: entry,
					})
				} else {
					fileEntries = append(fileEntries, PathEntry{
						Path:  path,
						Entry: entry,
					})
				}
			}
		}
		entryCh <- fileEntries
	}

	wg.Add(1)
	scanDir(dts.root)

	wg.Wait()
	close(entryCh)
	close(errCh)
}
