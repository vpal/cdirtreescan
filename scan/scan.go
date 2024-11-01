package scan

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"
)

type ScanAction int

const (
	List ScanAction = iota
	Count
)

type DirScanner struct {
	ctx         context.Context
	root        string
	concurrency uint64
}

func NewDirScanner(ctx context.Context, root string, concurrency uint64) *DirScanner {
	return &DirScanner{
		ctx:         ctx,
		root:        root,
		concurrency: concurrency,
	}
}

func (ds *DirScanner) Count() (errs []error) {
	var (
		resultWg  sync.WaitGroup
		dirCh     = make(chan string, ds.concurrency*2)
		fileCh    = make(chan string, ds.concurrency*2)
		errCh     = make(chan error, ds.concurrency*2)
		dirCount  int64
		fileCount int64
	)
	resultWg.Add(1)
	go func() {
		defer resultWg.Done()
		for err := range errCh {
			errs = append(errs, err)
		}
	}()

	resultWg.Add(1)
	go func() {
		defer resultWg.Done()
		for range dirCh {
			dirCount++
		}
	}()

	resultWg.Add(1)
	go func() {
		defer resultWg.Done()
		for range fileCh {
			fileCount++
		}
	}()

	ds.scanDirTree(dirCh, fileCh, errCh)

	resultWg.Wait()
	fmt.Printf("Number of directories: %v\n", dirCount)
	fmt.Printf("Number of files: %v\n", fileCount)
	return errs
}

func (ds *DirScanner) List() (errs []error) {
	var (
		resultWg sync.WaitGroup
		dirCh    = make(chan string, ds.concurrency*2)
		fileCh   = make(chan string, ds.concurrency*2)
		errCh    = make(chan error, ds.concurrency*2)
	)
	resultWg.Add(1)
	go func() {
		defer resultWg.Done()
		for err := range errCh {
			errs = append(errs, err)
		}
	}()

	resultWg.Add(1)
	go func() {
		defer resultWg.Done()
		for dir := range dirCh {
			fmt.Printf("d %v\n", dir)
		}
	}()

	resultWg.Add(1)
	go func() {
		defer resultWg.Done()
		for file := range fileCh {
			fmt.Printf("f %v\n", file)
		}
	}()

	ds.scanDirTree(dirCh, fileCh, errCh)

	resultWg.Wait()
	return errs
}

func (ds *DirScanner) scanDirTree(dirCh chan<- string, fileCh chan<- string, errCh chan<- error) {
	scanWg := sync.WaitGroup{}
	semCh := make(chan int, ds.concurrency)

	var scanDir func(string)
	scanDir = func(dir string) {
		defer scanWg.Done()
		defer func() {
			<-semCh
		}()

		dirCh <- dir
		entries, err := os.ReadDir(dir)
		if err != nil {
			errCh <- err
		}
		for _, entry := range entries {
			path := path.Join(dir, entry.Name())
			if entry.IsDir() {
				scanWg.Add(1)
				semCh <- 1
				go scanDir(path)
			} else {
				fileCh <- path
			}
		}
	}
	scanWg.Add(1)
	semCh <- 1
	scanDir(ds.root)

	scanWg.Wait()
	close(dirCh)
	close(fileCh)
	close(errCh)
}
