package handler

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/uuid"
)

type MultipartFileHandlerImpl struct {
}

func NewMultipartFileHandler() FileHandler {
	return &MultipartFileHandlerImpl{}
}

func (s *MultipartFileHandlerImpl) SaveFile(file File) (savedFile SavedFile, err error) {
	defer file.Close()

	err = os.MkdirAll("./uploads", os.ModePerm)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("Error Retrieving the File: %v", err)
		return nil, err
	}

	savedFilePath := "./uploads/" + strings.ReplaceAll(file.FileName(), ".csv", "_"+uuid.NewString()+".csv")
	dst, err := os.Create(savedFilePath)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("Error Retrieving the File: %v", err)
		return nil, err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file.Reader())
	if err != nil {
		fmt.Println(err)
		fmt.Printf("Error Retrieving the File: %v", err)
		return nil, err
	}

	savedOsFile, err := os.Open(savedFilePath)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("Error Retrieving the File: %v", err)
		return nil, err
	}

	return NewOsFile(savedOsFile), nil
}
