package bank_slip

import (
	"context"
	"fmt"
	"log"
	bankSlipEntities "performatic-file-processor/internal/bank_slip/entity"
	"performatic-file-processor/internal/infra/billing"
	emailService "performatic-file-processor/internal/infra/email"
	"performatic-file-processor/internal/messaging"
	"strings"
)

type ProcessBankSlipRowsServiceInterface interface {
	Execute(context context.Context, messagesChannel chan messaging.Message)
}

type ProcessBankSlipRowsService struct {
	bankSlipFileRepository bankSlipEntities.BankSlipFileMetadataRepository
	bankSlipRepository     bankSlipEntities.BankSlipRepository
	emailService           emailService.SendEmailService
	billingService         billing.BilingService
}

func NewProcessBankSlipRowsService(
	bankSlipFileRepository bankSlipEntities.BankSlipFileMetadataRepository,
	bankSlipRepository bankSlipEntities.BankSlipRepository,
) *ProcessBankSlipRowsService {
	return &ProcessBankSlipRowsService{
		bankSlipFileRepository: bankSlipFileRepository,
		bankSlipRepository:     bankSlipRepository,
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

			for row := range strings.SplitSeq(fileData, "\n") {
				if row == "" {
					continue
				}

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

			err = s.bankSlipRepository.InsertMany(bankSlips)

			s.processInsertedBankSlips(bankSlips)

			s.bankSlipRepository.UpdateMany(bankSlips)

			if err != nil {
				fmt.Printf("Error inserting new debts: %v\n", err.Error())
				continue
			}

			message.Commit()
			log.Printf("Inserted %d new debts\n", len(bankSlips))
		}
	}
}

func (s *ProcessBankSlipRowsService) processInsertedBankSlips(
	bankSlips map[bankSlipEntities.DebitId]*bankSlipEntities.BankSlip,
) (bankSlipsWithError map[bankSlipEntities.DebitId]*bankSlipEntities.BankSlip) {
	bankSlipsWithError = map[bankSlipEntities.DebitId]*bankSlipEntities.BankSlip{}

	errorsGeneratingBilling := *s.billingService.GenerateBiling(&bankSlips)
	for debtId := range bankSlips {
		bankSlipWithError := bankSlips[debtId]
		bankSlipWithError.ErrorGeneratingBilling(errorsGeneratingBilling[debtId].Error())
		if bankSlipWithError != nil {
			bankSlipsWithError[debtId] = bankSlipWithError
			delete(bankSlips, debtId)
		}
	}

	errorsSendingEmail := *s.emailService.SendBankSlipWaitingPaymentEmail(&bankSlips)
	for debtId := range bankSlips {
		bankSlipWithError := bankSlips[debtId]
		bankSlipWithError.ErrorSendingEmail(errorsSendingEmail[debtId].Error())
		if bankSlipWithError != nil {
			bankSlipsWithError[debtId] = bankSlipWithError
			delete(bankSlips, debtId)
		}
	}

	for _, bankSlip := range bankSlips {
		bankSlip.Success()
	}
	return bankSlipsWithError
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
