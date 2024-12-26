package scan

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"sync"
	"sync/atomic"
)

type DirTreeScanner struct {
	ctx         context.Context
	root        PathEntry
	chSize      int
	concurrency int
}

type PathEntry struct {
	Path  string
	Entry os.DirEntry
}

func NewDirTreeScanner(ctx context.Context, root string, concurrency int) (*DirTreeScanner, error) {
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
		chSize:      concurrency * 2,
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

func (dts *DirTreeScanner) ChSize() int {
	return dts.chSize
}

type TaskChannel struct {
	ch   chan PathEntry
	once sync.Once
}

func (tc *TaskChannel) Close() {
	tc.once.Do(func() {
		close(tc.ch)
	})
}

func (dts *DirTreeScanner) scanDirTree(entryCh chan<- []PathEntry, errCh chan<- error) {
	wg := sync.WaitGroup{}
	taskCh := &TaskChannel{
		ch:   make(chan PathEntry, dts.concurrency*100),
		once: sync.Once{},
	}
	taskCnt := atomic.Int32{}
	batchSize := 1024

	for i := 0; i < dts.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
		WLoop:
			for dir := range taskCh.ch {
				entryCh <- []PathEntry{
					{
						Path:  dir.Path,
						Entry: dir.Entry,
					},
				}

				file, err := os.Open(dir.Path)
				if err != nil {
					errCh <- err
					file.Close()
					continue WLoop
				}

			ELoop:
				for {
					entries, err := file.ReadDir(batchSize)
					if errors.Is(err, io.EOF) {
						break ELoop
					} else if err != nil {
						errCh <- err
						file.Close()
						continue WLoop
					}

					fileEntries := make([]PathEntry, 0, batchSize)
					for _, entry := range entries {
						path := path.Join(dir.Path, entry.Name())
						if entry.IsDir() {
							taskCnt.Add(1)
							taskCh.ch <- PathEntry{
								Path:  path,
								Entry: entry,
							}
						} else {
							fileEntries = append(fileEntries, PathEntry{
								Path:  path,
								Entry: entry,
							})
						}
					}
					entryCh <- fileEntries
				}
				file.Close()
				taskCnt.Add(-1)
				if taskCnt.Load() == 0 {
					taskCh.Close()
				}
			}
		}()
	}

	taskCnt.Add(1)
	taskCh.ch <- dts.root

	wg.Wait()
	close(entryCh)
	close(errCh)
}
