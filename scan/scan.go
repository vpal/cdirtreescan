package scan

import (
	"context"
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

func (dts *DirTreeScanner) Stream() (<-chan PathEntry, <-chan error) {
	var (
		entryCh = make(chan PathEntry, dts.chSize)
		errCh   = make(chan error, dts.chSize)
	)
	go dts.scanDirTree(entryCh, errCh)
	return entryCh, errCh
}

func (dts *DirTreeScanner) ChSize() int64 {
	return int64(dts.chSize)
}

func (dts *DirTreeScanner) scanDirTree(entryCh chan<- PathEntry, errCh chan<- error) {
	wg := sync.WaitGroup{}
	semCh := make(chan int, dts.concurrency)

	var scanDir func(PathEntry)
	scanDir = func(dir PathEntry) {
		defer wg.Done()
		semCh <- 1
		defer func() { <-semCh }()

		entryCh <- dir
		file, err := os.Open(dir.Path)
		if err != nil {
			errCh <- err
		}
		entries, err := file.ReadDir(0)
		if err != nil {
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
				entryCh <- PathEntry{
					Path:  path,
					Entry: entry,
				}
			}
		}
	}

	wg.Add(1)
	scanDir(dts.root)

	wg.Wait()
	close(entryCh)
	close(errCh)
}
