package filetype

import (
	"io/fs"

	"github.com/vpal/cdirtreescan/scan"
)

type FileType uint8

const (
	FileTypeRegular FileType = iota
	FileTypeBlockDevice
	FileTypeCharDevice
	FileTypeDirectory
	FileTypeSymlink
	FileTypeSocket
	FileTypeNamedPipe
	FileTypeOther
)

type FileTypeChar byte

const (
	FileTypeCharRegular     FileTypeChar = '-'
	FileTypeCharBlockDevice FileTypeChar = 'b'
	FileTypeCharCharDevice  FileTypeChar = 'c'
	FileTypeCharDirectory   FileTypeChar = 'd'
	FileTypeCharSymlink     FileTypeChar = 'l'
	FileTypeCharSocket      FileTypeChar = 'M'
	FileTypeCharNamedPipe   FileTypeChar = 'P'
	FileTypeCharOther       FileTypeChar = '?'
)

const (
	FileTypeDescriptionSingularRegular     string = "Regular file"
	FileTypeDescriptionSingularBlockDevice string = "Block device"
	FileTypeDescriptionSingularCharDevice  string = "Character device"
	FileTypeDescriptionSingularDirectory   string = "Directory"
	FileTypeDescriptionSingularSymlink     string = "Symbolic link"
	FileTypeDescriptionSingularSocket      string = "Socket"
	FileTypeDescriptionSingularNamedPipe   string = "FIFO (named pipe) file"
	FileTypeDescriptionSingularOther       string = "Other"
)

const (
	FileTypeDescriptionPluralRegular     string = "Regular files"
	FileTypeDescriptionPluralBlockDevice string = "Block devices"
	FileTypeDescriptionPluralCharDevice  string = "Character device file"
	FileTypeDescriptionPluralDirectory   string = "Directories"
	FileTypeDescriptionPluralSymlink     string = "Symbolic links"
	FileTypeDescriptionPluralSocket      string = "Sockets"
	FileTypeDescriptionPluralNamedPipe   string = "FIFOs (named pipe)"
	FileTypeDescriptionPluralOther       string = "Other"
)

var fileTypeIndicators = []FileTypeChar{
	FileTypeRegular:     FileTypeCharRegular,
	FileTypeBlockDevice: FileTypeCharBlockDevice,
	FileTypeCharDevice:  FileTypeCharCharDevice,
	FileTypeDirectory:   FileTypeCharDirectory,
	FileTypeSymlink:     FileTypeCharSymlink,
	FileTypeSocket:      FileTypeCharSocket,
	FileTypeNamedPipe:   FileTypeCharNamedPipe,
	FileTypeOther:       FileTypeCharOther,
}

var fileTypeDescriptionsSingular = []string{
	FileTypeRegular:     FileTypeDescriptionSingularRegular,
	FileTypeBlockDevice: FileTypeDescriptionSingularBlockDevice,
	FileTypeCharDevice:  FileTypeDescriptionSingularCharDevice,
	FileTypeDirectory:   FileTypeDescriptionSingularDirectory,
	FileTypeSymlink:     FileTypeDescriptionSingularSymlink,
	FileTypeSocket:      FileTypeDescriptionSingularSocket,
	FileTypeNamedPipe:   FileTypeDescriptionSingularNamedPipe,
	FileTypeOther:       FileTypeDescriptionSingularOther,
}

var fileTypeDescriptionsPlural = []string{
	FileTypeRegular:     FileTypeDescriptionPluralRegular,
	FileTypeBlockDevice: FileTypeDescriptionPluralBlockDevice,
	FileTypeCharDevice:  FileTypeDescriptionPluralCharDevice,
	FileTypeDirectory:   FileTypeDescriptionPluralDirectory,
	FileTypeSymlink:     FileTypeDescriptionPluralSymlink,
	FileTypeSocket:      FileTypeDescriptionPluralSocket,
	FileTypeNamedPipe:   FileTypeDescriptionPluralNamedPipe,
	FileTypeOther:       FileTypeDescriptionPluralOther,
}

func GetFileType(pathEntry scan.PathEntry) FileType {
	dirEntry := pathEntry.Entry
	switch mode := dirEntry.Type(); {
	case mode.IsDir():
		return FileTypeDirectory
	case mode.IsRegular():
		return FileTypeRegular
	case mode&fs.ModeSymlink != 0:
		return FileTypeSymlink
	case mode&fs.ModeDevice != 0:
		return FileTypeBlockDevice
	case mode&fs.ModeNamedPipe != 0:
		return FileTypeNamedPipe
	case mode&fs.ModeSocket != 0:
		return FileTypeSocket
	case mode&fs.ModeCharDevice != 0:
		return FileTypeCharDevice
	default:
		return FileTypeOther
	}
}

func GetFileTypeChar(fileType FileType) FileTypeChar {
	return fileTypeIndicators[fileType]
}

func GetFileTypeDescriptionSingular(fileType FileType) string {
	return fileTypeDescriptionsSingular[fileType]
}

func GetFileTypeDescriptionPlural(fileType FileType) string {
	return fileTypeDescriptionsPlural[fileType]
}
