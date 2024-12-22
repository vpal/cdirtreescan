package output

import (
	"fmt"
	"io"
	"sync"

	"github.com/vpal/cdirtreescan/filetype"
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

func (dtp *DirTreePrinter) PrintCount() error {
	var (
		wg     sync.WaitGroup
		errCnt uint64
		ftCnt  = []uint64{
			filetype.FileTypeRegular:     0,
			filetype.FileTypeBlockDevice: 0,
			filetype.FileTypeCharDevice:  0,
			filetype.FileTypeDirectory:   0,
			filetype.FileTypeSymlink:     0,
			filetype.FileTypeSocket:      0,
			filetype.FileTypeNamedPipe:   0,
			filetype.FileTypeOther:       0,
		}
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
				ftCnt[filetype.GetFileType(entry)]++
			}
		}
	}()

	wg.Wait()

	for t, c := range ftCnt {
		if c != 0 {
			ft := filetype.FileType(t)
			ftd := filetype.GetFileTypeDescriptionPlural(ft)
			fmt.Fprintf(dtp.writer, "%v: %v\n", ftd, c)
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
				fti := string(filetype.GetFileTypeIndicator(filetype.GetFileType(entry)))
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
