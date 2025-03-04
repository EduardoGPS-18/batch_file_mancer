package bank_slip

import (
	"context"
	"fmt"
	"log"
	bankSlipEntities "performatic-file-processor/internal/bank_slip/entity"
	bankSlipProviders "performatic-file-processor/internal/bank_slip/providers"
	"performatic-file-processor/internal/messaging"
	"strings"
)

type ProcessBankSlipRowsServiceInterface interface {
	Execute(context context.Context, messagesChannel chan messaging.Message)
}

type ProcessBankSlipRowsService struct {
	bankSlipFileRepository      bankSlipEntities.BankSlipFileMetadataRepository
	bankSlipRepository          bankSlipEntities.BankSlipRepository
	generateBillingAndSentEmail bankSlipProviders.GenerateBillingAndSentEmailProvider
}

func NewProcessBankSlipRowsService(
	bankSlipFileRepository bankSlipEntities.BankSlipFileMetadataRepository,
	bankSlipRepository bankSlipEntities.BankSlipRepository,
	generateBillingAndSentEmail bankSlipProviders.GenerateBillingAndSentEmailProvider,
) *ProcessBankSlipRowsService {
	return &ProcessBankSlipRowsService{
		bankSlipFileRepository:      bankSlipFileRepository,
		bankSlipRepository:          bankSlipRepository,
		generateBillingAndSentEmail: generateBillingAndSentEmail,
	}
}

func (s *ProcessBankSlipRowsService) Execute(context context.Context, messagesChannel chan messaging.Message) {
	select {
	case <-context.Done():
		log.Printf("Exiting ProcessBankSlipRowsService\n")
		return

	default:
		for message := range messagesChannel {

			fileData, fileHeader, fileId, err := s.getFieldsFromMessage(message)

			if err != nil {
				log.Printf("Error getting fields from message: %v\n", err)
				continue
			}

			bankSlips := map[bankSlipEntities.DebitId]*bankSlipEntities.BankSlip{}

			totalExpected := 0
			for row := range strings.SplitSeq(fileData, "\n") {
				if row == "" {
					continue
				}

				totalExpected++
				bankSlip, err := bankSlipEntities.NewBankSlipFromRow(fileId, row, fileHeader)
				if err != nil {
					fmt.Printf("Error creating Bank Slip Data: %v\n", err)
					continue
				}
				bankSlips[bankSlip.DebtId] = bankSlip
			}

			if len(bankSlips) <= 0 {
				fmt.Print("No new debts to insert\n")
				continue
			}

			insertedDebtIds, err := s.bankSlipRepository.InsertMany(&bankSlips)

			for debitId, success := range insertedDebtIds {
				if !success {
					delete(bankSlips, debitId)
				}
			}

			if err != nil {
				fmt.Printf("Error inserting new debts: %v\n", err.Error())
				continue
			}

			debitsWithErrors := s.generateBillingAndSentEmail.GenerateBillingAndSentEmail(&bankSlips)

			err = s.bankSlipRepository.UpdateMany(&bankSlips, debitsWithErrors)

			if err != nil {
				fmt.Printf("Error inserting new debts: %v\n", err.Error())
				continue
			}

			message.Commit()
			log.Printf("From %d inserted %d new debts\n", totalExpected, len(bankSlips))
		}
	}
}

func (s *ProcessBankSlipRowsService) getFieldsFromMessage(message messaging.Message) (fileData, fileHeader, fileId string, err error) {
	messageData, err := message.Data()
	if err != nil {
		return "", "", "", err
	}

	fileHeader = messageData["header"].(string)
	fileData = messageData["data"].(string)
	fileId = messageData["fileId"].(string)

	return fileData, fileHeader, fileId, nil
}
