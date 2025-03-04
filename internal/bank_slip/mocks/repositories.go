package bank_slip

import (
	"maps"
	entities "performatic-file-processor/internal/bank_slip/entity"

	"github.com/stretchr/testify/mock"
)

type BankSlipFileMetadataRepositoryMock struct {
	mock.Mock
}

func (m *BankSlipFileMetadataRepositoryMock) Insert(bankSlipFile *entities.BankSlipFileMetadata) error {
	args := m.Called(bankSlipFile)
	if args.Get(0) == nil {
		return nil
	}
	return args.Error(0)
}

type BankSlipRepositoryMock struct {
	mock.Mock
}

func (m *BankSlipRepositoryMock) InsertMany(bankSlips *entities.BankSlipMap) (map[entities.DebitId]entities.Success, error) {
	copy := make(entities.BankSlipMap)
	maps.Copy(copy, *bankSlips)
	args := m.Called(&copy)
	return args.Get(0).(map[entities.DebitId]entities.Success), args.Error(1)
}

func (m *BankSlipRepositoryMock) UpdateMany(bankSlips ...*entities.BankSlipMap) error {
	copies := make([]*entities.BankSlipMap, len(bankSlips))
	copy(copies, bankSlips)

	dynamicArgs := make([]interface{}, len(copies))
	for i, v := range copies {
		dynamicArgs[i] = v
	}

	args := m.Called(dynamicArgs...)
	return args.Error(0)
}
