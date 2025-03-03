package handler

import (
	"io"
	"mime/multipart"
)

type MultipartFile struct {
	file    multipart.File
	handler *multipart.FileHeader
}

func NewMultipartFile(file multipart.File, handler *multipart.FileHeader) *MultipartFile {
	return &MultipartFile{
		file:    file,
		handler: handler,
	}
}

func (s *MultipartFile) FileName() string {
	return s.handler.Filename
}

func (s *MultipartFile) Close() {
	s.file.Close()
}

func (s *MultipartFile) Reader() io.Reader {
	return s.file
}
