package output

import (
	"fmt"
	"io"
	"sync"

	"github.com/vpal/cdirtreescan/filetypes"
	"github.com/vpal/cdirtreescan/scan"
)

func NewDirTreePrinter(dts *scan.DirTreeScanner, writer io.Writer, errWriter io.Writer, displayErrors bool) *DirTreePrinter {
	return &DirTreePrinter{
		dts:           dts,
		writer:        writer,
		errWriter:     errWriter,
		displayErrors: displayErrors,
	}
}

type DirTreePrinter struct {
	dts           *scan.DirTreeScanner
	writer        io.Writer
	errWriter     io.Writer
	displayErrors bool
}

type fileTypeCounter struct {
	counts []uint64
}

func (ftc *fileTypeCounter) Inc(ft filetypes.FileType) {
	ftc.counts[ft]++
}

func (ftc *fileTypeCounter) Get(ft filetypes.FileType) uint64 {
	return ftc.counts[ft]
}

func (ftc *fileTypeCounter) Total() uint64 {
	var total uint64
	for _, cnt := range ftc.counts {
		total += cnt
	}
	return total
}

func newFileTypeCounter() *fileTypeCounter {
	return &fileTypeCounter{
		counts: make([]uint64, len(filetypes.FileTypes)),
	}
}

func (dtp *DirTreePrinter) PrintCount() error {
	var (
		wg     sync.WaitGroup
		errCnt uint64
		ftCnt  = newFileTypeCounter()
	)

	entryCh, errCh := dtp.dts.Stream()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range errCh {
			errCnt++
			if dtp.displayErrors {
				fmt.Fprintln(dtp.errWriter, err)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for entries := range entryCh {
			for _, entry := range entries {
				ftCnt.Inc(filetypes.GetFileType(entry))
			}
		}
	}()

	wg.Wait()

	fmt.Fprintf(dtp.writer, "Total: %v\n", ftCnt.Total())
	for _, ft := range filetypes.FileTypes {
		cnt := ftCnt.Get(ft)
		if cnt != 0 {
			ftDesc := filetypes.GetFileTypeDescriptionPlural(ft)
			fmt.Fprintf(dtp.writer, "%v: %v\n", ftDesc, cnt)
		}
	}

	if errCnt != 0 {
		return fmt.Errorf("%v error(s) happened during scanning", errCnt)
	}
	return nil
}

func (dtp *DirTreePrinter) PrintList() error {
	var (
		wg     sync.WaitGroup
		errCnt uint64
	)

	entryCh, errCh := dtp.dts.Stream()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range errCh {
			errCnt++
			if dtp.displayErrors {
				fmt.Fprintln(dtp.errWriter, err)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for entries := range entryCh {
			for _, entry := range entries {
				fti := string(filetypes.GetFileTypeIndicator(filetypes.GetFileType(entry)))
				fmt.Fprintf(dtp.writer, "%v %v\n", fti, entry.Path)
			}
		}
	}()

	wg.Wait()

	if errCnt != 0 {
		return fmt.Errorf("%v error(s) happened during scanning", errCnt)
	}
	return nil
}
