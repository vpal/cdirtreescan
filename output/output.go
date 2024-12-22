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

func (dtp *DirTreePrinter) PrintCount() error {
	var (
		wg     sync.WaitGroup
		errCnt uint64
		ftCnt  = []uint64{
			filetypes.FileTypeRegular:     0,
			filetypes.FileTypeBlockDevice: 0,
			filetypes.FileTypeCharDevice:  0,
			filetypes.FileTypeDirectory:   0,
			filetypes.FileTypeSymlink:     0,
			filetypes.FileTypeSocket:      0,
			filetypes.FileTypeNamedPipe:   0,
			filetypes.FileTypeOther:       0,
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
				ftCnt[filetypes.GetFileType(entry)]++
			}
		}
	}()

	wg.Wait()

	for t, c := range ftCnt {
		if c != 0 {
			ft := filetypes.FileType(t)
			ftd := filetypes.GetFileTypeDescriptionPlural(ft)
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
