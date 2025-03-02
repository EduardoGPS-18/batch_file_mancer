package bank_slip

import "time"

type BankSlipFileMetadataRepository interface {
	Insert(bankSlipFile *BankSlipFileMetadata) error
}

type BankSlipFileMetadata struct {
	ID        string
	FileName  string
	CreatedAt time.Time
}

func NewBankSlipFileMetadata(fileName string) *BankSlipFileMetadata {
	return &BankSlipFileMetadata{
		FileName: fileName,
	}
}
