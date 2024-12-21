package output

import (
	"fmt"
	"io"
	"sync"

	"github.com/vpal/cdirtreescan/filetype"
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
		wg    sync.WaitGroup
		errs  []error
		ftCnt = []uint64{
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
			errs = append(errs, err)
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
		for entries := range entryCh {
			for _, entry := range entries {
				fti := string(filetype.GetFileTypeChar(filetype.GetFileType(entry)))
				fmt.Fprintf(dtp.writer, "%v %v\n", fti, entry.Path)
			}
		}
	}()

	wg.Wait()
}
