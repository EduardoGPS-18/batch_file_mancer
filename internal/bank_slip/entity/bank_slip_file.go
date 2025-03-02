package bank_slip

import "time"

type BankSlipFileMetadataRepository interface {
	Insert(bankSlipFile *BankSlipFileMetadata) error
}

type BankSlipFileMetadata struct {
	ID        int
	FileName  string
	CreatedAt time.Time
}

func NewBankSlipFile(fileName string) *BankSlipFileMetadata {
	return &BankSlipFileMetadata{
		FileName: fileName,
	}
}
