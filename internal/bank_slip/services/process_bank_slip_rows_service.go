package bank_slip

import (
	"fmt"
	"log"
	bankSlipEntities "performatic-file-processor/internal/bank_slip/entity"
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

func (s *ProcessBankSlipRowsService) Execute(messagesChannel chan map[string]any) {
	for message := range messagesChannel {
		fileData, fileHeader, fileId := s.getFieldsFromMessage(message)

		bankSlipList := map[bankSlipEntities.DebitId]*bankSlipEntities.BankSlip{}
		debitIds := []string{}

		inserted := 0
		for row := range strings.SplitSeq(fileData, "\n") {
			if row == "" {
				continue
			}

			bankSlip, err := bankSlipEntities.NewBankSlipFromRow(fileId, row, fileHeader)
			if err != nil {
				fmt.Printf("Error creating Bank Slip Data: %v\n", err)
				continue
			}
			bankSlipList[bankSlip.DebtId] = bankSlip
			debitIds = append(debitIds, fmt.Sprintf("'%s'", bankSlip.DebtId))
			inserted++
		}

		alreadyExistingDebts, err := s.bankSlipRepository.GetExistingByDebitIds(debitIds)
		if err != nil {
			fmt.Printf("Error getting existing debts: %v\n", err)
			continue
		}

		for existingDebit := range alreadyExistingDebts {
			bankSlipList[existingDebit].SetRowWithError("Debt already exists")
		}

		if len(bankSlipList) <= 0 {
			fmt.Print("No new debts to insert\n")
			continue
		}

		err = s.bankSlipRepository.InsertMany(bankSlipList)
		if err != nil {
			fmt.Printf("Error inserting new debts: %v\n", err)
			continue
		}
		log.Printf("Inserted %d new debts\n", inserted)
	}
}
func (s *ProcessBankSlipRowsService) getFieldsFromMessage(message map[string]any) (fileData, fileHeader, fileId string) {

	fileHeader = message["header"].(string)
	fileData = message["data"].(string)
	fileId = message["fileId"].(string)

	return fileData, fileHeader, fileId
}
