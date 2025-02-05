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

	"github.com/vpal/cdirtreescan/dcbuffer"
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

type workerSync struct {
	inCh    chan PathEntry
	outCh   chan PathEntry
	buffer  *dcbuffer.DynamicCircularBuffer[PathEntry]
	taskCnt atomic.Int32
}

func (ws *workerSync) incTaskCnt() {
	ws.taskCnt.Add(1)
}

func (ws *workerSync) decTaskCnt() {
	ws.taskCnt.Add(-1)
}

func (ws *workerSync) getTaskCnt() int32 {
	return ws.taskCnt.Load()
}

func (dts *DirTreeScanner) scanDirTree(entryCh chan<- []PathEntry, errCh chan<- error) {
	wg := sync.WaitGroup{}
	ws := &workerSync{
		inCh:   make(chan PathEntry, dts.concurrency*2),
		outCh:  make(chan PathEntry, dts.concurrency*2),
		buffer: dcbuffer.NewDynamicCircularBuffer[PathEntry](dts.concurrency),
	}
	batchSize := 1024

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if !ws.buffer.Empty() {
				entry, _ := ws.buffer.Peek()
				select {
				case ws.inCh <- entry:
					ws.buffer.Dequeue()
				case entry, ok := <-ws.outCh:
					if !ok {
						close(ws.inCh)
						return
					}
					ws.buffer.Enqueue(entry)
				}
			} else {
				entry, ok := <-ws.outCh
				if !ok {
					close(ws.inCh)
					return
				}
				ws.buffer.Enqueue(entry)
			}
		}
	}()

	for i := 0; i < dts.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

		WorkerLoop:
			for dir := range ws.inCh {
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
					continue WorkerLoop
				}

			EntryLoop:
				for {
					entries, err := file.ReadDir(batchSize)
					if errors.Is(err, io.EOF) {
						break EntryLoop
					} else if err != nil {
						errCh <- err
						file.Close()
						continue WorkerLoop
					}

					fileEntries := make([]PathEntry, 0, len(entries))
					for _, entry := range entries {
						path := path.Join(dir.Path, entry.Name())
						if entry.IsDir() {
							ws.incTaskCnt()
							ws.outCh <- PathEntry{
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
				ws.decTaskCnt()
				if ws.getTaskCnt() == 0 {
					close(ws.outCh)
				}
			}
		}()
	}

	ws.incTaskCnt()
	ws.inCh <- dts.root

	wg.Wait()
	close(entryCh)
	close(errCh)
}
