package bank_slip

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strings"

	entities "performatic-file-processor/internal/bank_slip/entity"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type ReceiveUploadService struct {
	bankSlipRepository             entities.BankSlipRepository
	bankSlipFileMetadataRepository entities.BankSlipFileMetadataRepository
}

func NewReceiveUploadService(bankSlipRepo entities.BankSlipRepository, bankSlipFileRepo entities.BankSlipFileMetadataRepository) *ReceiveUploadService {
	return &ReceiveUploadService{
		bankSlipRepository:             bankSlipRepo,
		bankSlipFileMetadataRepository: bankSlipFileRepo,
	}
}

func (s *ReceiveUploadService) Execute(file multipart.File, fileName string) {
	messagesChannel := make(chan *kafka.Message)

	go func() {
		for message := range messagesChannel {
			fileData, fileHeader, fileId, error := s.getFieldsFromMessage(message)
			if error != nil || fileData == "" {
				switch {
				case error != nil:
					fmt.Printf("Error: data or header is empty\n")
				case fileData == "":
					fmt.Printf("Error getting file fields: %v\n", error)
				}
				continue
			}

			bankSlipList := map[entities.DebitId]*entities.BankSlip{}
			debitIds := []string{}

			for row := range strings.SplitSeq(fileData, "\n") {
				if row == "" {
					continue
				}

				bankSlip, err := entities.NewBankSlipFromRow(fileId, row, fileHeader)
				if err != nil {
					fmt.Printf("Error creating Bank Slip Data: %v\n", err)
					continue
				}
				bankSlipList[bankSlip.DebtId] = bankSlip
				debitIds = append(debitIds, fmt.Sprintf("'%s'", bankSlip.DebtId))
			}

			alreadyExistingDebts, err := s.bankSlipRepository.GetExistingByDebitIds(debitIds)
			if err != nil {
				fmt.Printf("Error getting existing debts: %v\n", err)
				continue
			}

			for existingDebit := range alreadyExistingDebts {
				bankSlipList[existingDebit].SetRowWithError("Debt already exists")
			}

			if len(alreadyExistingDebts) <= 0 {
				fmt.Print("No new debts to insert\n")
				continue
			}

			err = s.bankSlipRepository.InsertMany(bankSlipList)
			if err != nil {
				fmt.Printf("Error inserting new debts: %v\n", err)
				continue
			}
		}
	}()
}

func (s *ReceiveUploadService) getFieldsFromMessage(message *kafka.Message) (fileData, fileHeader, fileId string, err error) {
	jsonData := map[string]string{}
	err = json.Unmarshal(message.Value, &jsonData)
	if err != nil {
		fmt.Printf("Error Unmarshalling Message: %v\n", err)
		return "", "", "", err
	}
	fileHeader = jsonData["header"]
	fileData = jsonData["data"]
	fileId = jsonData["fileId"]

	return fileData, fileHeader, fileId, nil
}
