package output

import (
	"fmt"
	"io"
	"sync"

	"github.com/vpal/cdirtreescan/scan"
)

func NewDirTreePrinter(dts *scan.DirTreeScanner, writer io.Writer, errWriter io.Writer) *DirTreePrinter {
	return &DirTreePrinter{
		dts:    dts,
		writer: writer,
	}
}

type DirTreePrinter struct {
	dts    *scan.DirTreeScanner
	writer io.Writer
}

func (dtp *DirTreePrinter) PrintCount() {
	var (
		wg        sync.WaitGroup
		errs      []error
		dirCount  uint64
		fileCount uint64
	)

	entryCh, errCh := dtp.dts.Stream()

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
		for entry := range entryCh {
			switch {
			case entry.Entry.IsDir():
				dirCount++
			default:
				fileCount++
			}
		}
	}()

	wg.Wait()
	fmt.Printf("Number of directories: %v\n", dirCount)
	fmt.Printf("Number of files: %v\n", fileCount)
}

func (dtp *DirTreePrinter) PrintList() {
	var (
		wg   sync.WaitGroup
		errs []error
	)

	entryCh, errCh := dtp.dts.Stream()

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
		for entry := range entryCh {
			mode := entry.Entry.Type().String()
			mode = mode[:len(mode)-9]
			fmt.Fprintf(dtp.writer, "%v %v\n", mode, entry.Path)
		}
	}()

	wg.Wait()
}
