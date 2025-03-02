package bank_slip

import (
	"fmt"
	"log"
	bankSlipEntities "performatic-file-processor/internal/bank_slip/entity"
	"performatic-file-processor/internal/messaging"
	"strings"
)

type ProcessBankSlipRowsService struct {
	bankSlipFileRepository bankSlipEntities.BankSlipFileMetadataRepository
	bankSlipRepository     bankSlipEntities.BankSlipRepository
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

func (s *ProcessBankSlipRowsService) Execute(messagesChannel chan messaging.Message) {
	for message := range messagesChannel {
		fileData, fileHeader, fileId, err := s.getFieldsFromMessage(message)

		if err != nil {
			log.Printf("Error getting fields from message: %v\n", err)
			continue
		}

		bankSlips := map[bankSlipEntities.DebitId]*bankSlipEntities.BankSlip{}
		debitIds := []string{}

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
			debitIds = append(debitIds, fmt.Sprintf("'%s'", bankSlip.DebtId))
		}

		alreadyExistingDebts, err := s.bankSlipRepository.GetExistingByDebitIds(debitIds)
		if err != nil {
			fmt.Printf("Error getting existing debts: %v\n", err)
			continue
		}

		for existingDebit := range alreadyExistingDebts {
			bankSlips[existingDebit].UpdateRowToError("Debt already exists")
		}

		if len(bankSlips) <= 0 {
			fmt.Print("No new debts to insert\n")
			continue
		}

		err = s.bankSlipRepository.InsertMany(bankSlips)
		if err != nil {
			fmt.Printf("Error inserting new debts: %v\n", err.Error())
			continue
		}

		message.Commit()
		log.Printf("Inserted %d new debts\n", len(bankSlips))
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
