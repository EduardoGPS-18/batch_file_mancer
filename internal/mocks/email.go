package mocks

import (
	"maps"
	bankSlipEntities "performatic-file-processor/internal/bank_slip/entity"

	"github.com/stretchr/testify/mock"
)

type EmailServicesMock struct {
	mock.Mock
}

func (m *EmailServicesMock) SendBankSlipWaitingPaymentEmail(
	bankSlip *bankSlipEntities.BankSlipMap,
) *map[bankSlipEntities.DebitId]error {
	bankSlipCopy := make(bankSlipEntities.BankSlipMap)
	maps.Copy(bankSlipCopy, *bankSlip)
	args := m.Called(&bankSlipCopy)
	return args.Get(0).(*map[bankSlipEntities.DebitId]error)
}
