package scan

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"
)

type DirScanner struct {
	ctx         context.Context
	root        string
	resChSize   uint64
	concurrency uint64
}

func NewDirScanner(ctx context.Context, root string, concurrency uint64) *DirScanner {
	return &DirScanner{
		ctx:         ctx,
		root:        root,
		concurrency: concurrency,
		resChSize:   concurrency * 2,
	}
}

func (ds *DirScanner) Count() (errs []error) {
	var (
		wg        sync.WaitGroup
		dirCh     = make(chan string, ds.resChSize)
		fileCh    = make(chan string, ds.resChSize)
		errCh     = make(chan error, ds.resChSize)
		dirCount  int64
		fileCount int64
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range errCh {
			errs = append(errs, err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for range dirCh {
			dirCount++
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for range fileCh {
			fileCount++
		}
	}()

	ds.scanDir(dirCh, fileCh, errCh)

	wg.Wait()
	fmt.Printf("Number of directories: %v\n", dirCount)
	fmt.Printf("Number of files: %v\n", fileCount)
	return errs
}

func (ds *DirScanner) List() (errs []error) {
	var (
		wg     sync.WaitGroup
		dirCh  = make(chan string, ds.resChSize)
		fileCh = make(chan string, ds.resChSize)
		errCh  = make(chan error, ds.resChSize)
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range errCh {
			errs = append(errs, err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for dir := range dirCh {
			fmt.Printf("d %v\n", dir)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for file := range fileCh {
			fmt.Printf("f %v\n", file)
		}
	}()

	ds.scanDir(dirCh, fileCh, errCh)

	wg.Wait()
	return errs
}

func (ds *DirScanner) scanDir(dirCh chan<- string, fileCh chan<- string, errCh chan<- error) {
	wg := sync.WaitGroup{}
	semCh := make(chan int, ds.concurrency)

	var scanDir func(string)
	scanDir = func(dir string) {
		defer wg.Done()

		semCh <- 1
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
				wg.Add(1)
				go scanDir(path)
			} else {
				fileCh <- path
			}
		}
	}

	wg.Add(1)
	scanDir(ds.root)

	wg.Wait()
	close(dirCh)
	close(fileCh)
	close(errCh)
}
