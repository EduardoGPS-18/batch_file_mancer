package bank_slip

import (
	entities "performatic-file-processor/internal/bank_slip/entity"

	"github.com/stretchr/testify/mock"
)

type BankSlipFileMetadataRepositoryMock struct {
	mock.Mock
}

func (m *BankSlipFileMetadataRepositoryMock) Insert(bankSlipFile *entities.BankSlipFileMetadata) error {
	args := m.Called(bankSlipFile)
	return args.Error(0)
}

type BankSlipRepositoryMock struct {
	mock.Mock
}

func (m *BankSlipRepositoryMock) GetExistingByDebitIds(debitIds []string) (map[entities.DebitId]entities.Existing, error) {
	args := m.Called(debitIds)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(map[entities.DebitId]entities.Existing), args.Error(1)
}

func (m *BankSlipRepositoryMock) InsertMany(bankSlips map[entities.DebitId]*entities.BankSlip) error {
	args := m.Called(bankSlips)
	return args.Error(0)
}
