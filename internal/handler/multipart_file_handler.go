package handler

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strings"

	"github.com/google/uuid"
)

type MultipartFileHandler interface {
	SaveMultipartFileLocally(file multipart.File, fileName string) (filePath *string, deleteFile func(), err error)
}

type MultipartFileHandlerImpl struct {
}

func NewMultipartFileHandler() MultipartFileHandler {
	return &MultipartFileHandlerImpl{}
}

func (s *MultipartFileHandlerImpl) SaveMultipartFileLocally(file multipart.File, fileName string) (filePath *string, deleteFile func(), err error) {
	// file, handler, err := req.FormFile("file")
	// if err != nil {
	// 	fmt.Println("Error Retrieving the File")
	// 	fmt.Printf("Error Retrieving the File: %v", err)
	// 	return nil, nil, err
	// }

	defer file.Close()

	err = os.MkdirAll("./uploads", os.ModePerm)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("Error Retrieving the File: %v", err)
		return nil, nil, err
	}

	savedFilePath := "./uploads/" + strings.ReplaceAll(fileName, ".csv", "_"+uuid.NewString()+".csv")
	dst, err := os.Create(savedFilePath)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("Error Retrieving the File: %v", err)
		return nil, nil, err
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("Error Retrieving the File: %v", err)
		return nil, nil, err
	}

	deleteFile = func() {
		err := os.Remove(savedFilePath)
		if err != nil {
			fmt.Println(err)
			fmt.Printf("Error Retrieving the File: %v", err)
		}
	}

	return &savedFilePath, deleteFile, nil
}
