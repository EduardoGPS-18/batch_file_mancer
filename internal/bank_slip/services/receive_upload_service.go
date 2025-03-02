package bank_slip

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"strings"
	"sync"
	"time"

	bankSlipEntities "performatic-file-processor/internal/bank_slip/entity"
	"performatic-file-processor/internal/handler"
	"performatic-file-processor/internal/messaging"
)

type Row struct {
	data   []byte
	header string
}

type ReceiveUploadService struct {
	bankSlipRepository             bankSlipEntities.BankSlipRepository
	bankSlipFileMetadataRepository bankSlipEntities.BankSlipFileMetadataRepository
	multipartFileHandler           handler.MultipartFileHandler
	messageProducer                messaging.MessageProducer
	workers                        int
	bufferSize                     int
}

func NewReceiveUploadService(
	bankSlipRepo bankSlipEntities.BankSlipRepository,
	bankSlipFileRepo bankSlipEntities.BankSlipFileMetadataRepository,
	multipartFileHandler handler.MultipartFileHandler,
	messageProducer messaging.MessageProducer,
	bufferSize int,
	workers int,
) *ReceiveUploadService {
	return &ReceiveUploadService{
		bankSlipRepository:             bankSlipRepo,
		bankSlipFileMetadataRepository: bankSlipFileRepo,
		multipartFileHandler:           multipartFileHandler,
		messageProducer:                messageProducer,
		workers:                        workers,
		bufferSize:                     bufferSize,
	}
}

func (s *ReceiveUploadService) Execute(file multipart.File, fileHeader *multipart.FileHeader) error {
	start := time.Now()

	bankSlipFile := bankSlipEntities.NewBankSlipFileMetadata(fileHeader.Filename)

	err := s.bankSlipFileMetadataRepository.Insert(bankSlipFile)
	if err != nil {
		return err
	}

	filePath, deleteFile, err := s.multipartFileHandler.SaveMultipartFileLocally(file, fileHeader.Filename)
	if err != nil {
		return err
	}
	defer deleteFile()

	buffer := make([]byte, s.bufferSize)
	fileChannel := make(chan Row, s.workers)

	var wg sync.WaitGroup
	wg.Add(s.workers)

	for i := range s.workers {
		go s.processFile(i, fileChannel, bankSlipFile.ID, &wg)
	}

	locallyFile, err := os.Open(*filePath)
	if err != nil {
		fmt.Printf("Error Opening the File: %v", err)
		return err
	}
	defer locallyFile.Close()

	isFirstItem := true
	var remainder string
	var header string
	for {
		bytesRead, err := locallyFile.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error Reading the File: %v", err)
			}
			break
		}

		str := string(buffer[:bytesRead])

		lastItemIndex := strings.LastIndex(str, "\n")
		if lastItemIndex == -1 {
			lastItemIndex = len(str)
		}
		if !isFirstItem {
			fullRowsValid := remainder + str[:lastItemIndex]
			fileChannel <- Row{data: []byte(fullRowsValid), header: header}
		} else {
			isFirstItem = false
			firstItem := strings.Index(str, "\n")
			header = str[:firstItem]
			rowsWithoutHeader := str[firstItem+1 : lastItemIndex]
			fileChannel <- Row{data: []byte(rowsWithoutHeader), header: header}
		}
		remainder = str[lastItemIndex:]
	}
	close(fileChannel)

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Printf("Time taken: %s\n", elapsed)
	return nil
}

func (f *ReceiveUploadService) processFile(worker int, fileChannel chan Row, fileId string, wg *sync.WaitGroup) {
	for {
		row, ok := <-fileChannel
		if len(row.data) == 0 || !ok {
			fmt.Printf("Worker %d finished processing\n", worker)
			break
		}

		message := map[string]any{"data": string(row.data), "header": row.header, "fileId": fileId}

		err := f.messageProducer.Publish(context.TODO(), "rows-to-process", message)
		if err != nil {
			fmt.Print("Error posting message", err)
			break
		}
	}

	wg.Done()
}
