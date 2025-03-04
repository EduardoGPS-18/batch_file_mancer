package handler

import "io"

type File interface {
	FileName() string
	Close()
	Reader() io.Reader
}

type SavedFile interface {
	Delete()
	Open() io.Reader
	Filepath() string
}

type FileHandler interface {
	SaveFile(file File) (savedFile SavedFile, err error)
}
