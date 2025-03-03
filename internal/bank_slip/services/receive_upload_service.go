package bank_slip

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
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
	fileHandler                    handler.FileHandler
	messageProducer                messaging.MessageProducer
	workers                        int
	bufferSize                     int
}

func NewReceiveUploadService(
	bankSlipRepo bankSlipEntities.BankSlipRepository,
	bankSlipFileRepo bankSlipEntities.BankSlipFileMetadataRepository,
	multipartFileHandler handler.FileHandler,
	messageProducer messaging.MessageProducer,
	bufferSize int,
	workers int,
) *ReceiveUploadService {
	return &ReceiveUploadService{
		bankSlipRepository:             bankSlipRepo,
		bankSlipFileMetadataRepository: bankSlipFileRepo,
		fileHandler:                    multipartFileHandler,
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

	savedFile, err := s.fileHandler.SaveFile(handler.NewMultipartFile(file, fileHeader))
	if err != nil {
		return err
	}
	defer savedFile.Delete()

	buffer := make([]byte, s.bufferSize)
	fileChannel := make(chan Row, s.workers)

	var wg sync.WaitGroup
	wg.Add(s.workers)

	for i := range s.workers {
		go s.processFile(i, fileChannel, bankSlipFile.ID, &wg)
	}

	locallyFile := savedFile.Open()

	header, remainder := s.readFileHeader(locallyFile, buffer)

	if header == "" {
		close(fileChannel)
		elapsed := time.Since(start)
		fmt.Printf("Time taken: %s\n", elapsed)
		return errors.New("header not found")
	}

	s.readFileContent(locallyFile, buffer, remainder, fileChannel, header)

	close(fileChannel)

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Printf("Time taken: %s\n", elapsed)
	return nil
}

func (*ReceiveUploadService) readFileContent(locallyFile io.Reader, buffer []byte, headerRemaining string, fileChannel chan Row, header string) {
	remainder := headerRemaining
	for {
		bytesRead, err := locallyFile.Read(buffer)
		if bytesRead == 0 {
			break
		}
		if err != nil {
			log.Printf("Error Reading the File: %v", err)
			break
		}

		str := string(buffer)

		lastItemIndex := strings.LastIndex(str, "\n")
		lastRemainingIndex := strings.LastIndex(remainder, "\n")
		if lastItemIndex == -1 && lastRemainingIndex == -1 {
			remainder = remainder + str
			continue
		}

		fullRowsValid := remainder + str[:lastItemIndex]
		if lastItemIndex == 0 {
			fullRowsValid = remainder + str
		}
		fileChannel <- Row{data: []byte(fullRowsValid), header: header}
		remainder = str[lastItemIndex+1:]
	}
}

func (*ReceiveUploadService) readFileHeader(locallyFile io.Reader, buffer []byte) (string, string) {
	var header string
	var remainder string
	for {
		bytesRead, err := locallyFile.Read(buffer)
		if bytesRead == 0 {
			break
		}
		if err != nil {
			log.Printf("Error Reading the File: %v", err)
			break
		}
		str := string(buffer)

		finalHeaderIdx := strings.Index(str, "\n")
		if finalHeaderIdx == -1 {
			header += str
			continue
		}
		header += str[:finalHeaderIdx]
		if finalHeaderIdx != -1 {
			remainder = str[finalHeaderIdx+1:]
			break
		}
	}
	return header, remainder
}

func (f *ReceiveUploadService) processFile(worker int, fileChannel chan Row, fileId string, wg *sync.WaitGroup) {
	for {
		row, ok := <-fileChannel
		if len(row.data) == 0 && !ok {
			log.Printf("Worker %d finished processing\n", worker)
			break
		}

		message := map[string]any{"data": string(row.data), "header": row.header, "fileId": fileId}

		err := f.messageProducer.Publish(context.TODO(), "rows-to-process", message)
		if err != nil {
			log.Print("Error posting message", err)
			continue
		}
	}

	wg.Done()
}
