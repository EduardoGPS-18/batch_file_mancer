package handler

import (
	"io"
	"os"
)

type OsFile struct {
	file *os.File
}

func NewOsFile(file *os.File) *OsFile {
	return &OsFile{
		file: file,
	}
}

func (s *OsFile) Delete() {
	s.file.Close()
	os.Remove(s.file.Name())
}

func (s *OsFile) Open() io.Reader {
	return s.file
}
