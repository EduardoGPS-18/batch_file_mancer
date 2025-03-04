package bank_slip

import (
	"maps"
	bankSlipEntities "performatic-file-processor/internal/bank_slip/entity"

	"github.com/stretchr/testify/mock"
)

type GenerateBillingAndSentEmailProviderMock struct {
	mock.Mock
}

func (m *GenerateBillingAndSentEmailProviderMock) GenerateBillingAndSentEmail(
	bankSlips *bankSlipEntities.BankSlipMap,
) *bankSlipEntities.BankSlipMap {
	copy := make(bankSlipEntities.BankSlipMap)
	maps.Copy(copy, *bankSlips)
	args := m.Called(&copy)
	return args.Get(0).(*bankSlipEntities.BankSlipMap)
}
