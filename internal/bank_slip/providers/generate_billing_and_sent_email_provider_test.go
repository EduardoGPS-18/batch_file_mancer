package bank_slip

import (
	bsEntities "performatic-file-processor/internal/bank_slip/entity"
	"performatic-file-processor/internal/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type GenerateBillingAndSentEmailProviderTestSuite struct {
	suite.Suite
	mockEmailService   *mocks.EmailServicesMock
	mockBillingService *mocks.BillingServicesMock
	provider           GenerateBillingAndSentEmailProvider
}

func (s *GenerateBillingAndSentEmailProviderTestSuite) SetupTest() {
	s.mockEmailService = new(mocks.EmailServicesMock)
	s.mockBillingService = new(mocks.BillingServicesMock)
	s.provider = NewGenerateBillingAndSentEmailProvider(
		s.mockEmailService,
		s.mockBillingService,
	)
}

func TestNewGenerateBillingAndSentEmailProvider(t *testing.T) {
	suite.Run(t, new(GenerateBillingAndSentEmailProviderTestSuite))
}

func (s *GenerateBillingAndSentEmailProviderTestSuite) TestGenerateBillingAndSentEmailProvider_ShouldCallGenerateForEachBankSlip() {
	bankSlips := &bsEntities.BankSlipMap{
		"debit1": &bsEntities.BankSlip{
			DebtId: "debit1",
		},
		"debit2": &bsEntities.BankSlip{
			DebtId: "debit2",
		},
	}

	s.mockBillingService.On("GenerateBiling", bankSlips).Return(&map[bsEntities.DebitId]error{}).Once()
	s.mockEmailService.On("SendBankSlipWaitingPaymentEmail", bankSlips).Return(&map[bsEntities.DebitId]error{}).Once()

	s.provider.GenerateBillingAndSentEmail(bankSlips)

	s.mockBillingService.AssertCalled(s.T(), "GenerateBiling", mock.MatchedBy(func(bankSlips *bsEntities.BankSlipMap) bool {
		return (*bankSlips)["debit1"].DebtId == "debit1" && (*bankSlips)["debit2"].DebtId == "debit2"
	}))
}

func (s *GenerateBillingAndSentEmailProviderTestSuite) TestGenerateBillingAndSentEmailProvider_ShouldCallSendEmailToOnlySucceedsOnGenerated() {
	bankSlips := &bsEntities.BankSlipMap{
		"debit1": &bsEntities.BankSlip{
			DebtId: "debit1",
		},
		"debit2": &bsEntities.BankSlip{
			DebtId: "debit2",
		},
	}

	s.mockBillingService.On("GenerateBiling", mock.Anything).
		Return(&map[bsEntities.DebitId]error{"debit1": assert.AnError}).
		Once()
	s.mockEmailService.On("SendBankSlipWaitingPaymentEmail", mock.Anything).
		Return(&map[bsEntities.DebitId]error{}).
		Once()

	rowsWithError := s.provider.GenerateBillingAndSentEmail(bankSlips)

	s.mockBillingService.AssertCalled(s.T(), "GenerateBiling", mock.MatchedBy(func(bankSlips *bsEntities.BankSlipMap) bool {
		return (*bankSlips)["debit1"].DebtId == "debit1" && (*bankSlips)["debit2"].DebtId == "debit2"
	}))
	s.mockEmailService.AssertCalled(s.T(), "SendBankSlipWaitingPaymentEmail", mock.MatchedBy(func(bankSlips *bsEntities.BankSlipMap) bool {
		return len(*bankSlips) == 1 && (*bankSlips)["debit2"].DebtId == "debit2"
	}))
	assert.Equal(s.T(), len(*bankSlips), 1)
	assert.Equal(s.T(), bsEntities.BankSlipStatusSuccess, (*bankSlips)["debit2"].Status)
	assert.Equal(s.T(), len(*rowsWithError), 1)
	assert.Equal(s.T(), bsEntities.BankSlipStatusGenerateBillingError, (*rowsWithError)["debit1"].Status)
}

func (s *GenerateBillingAndSentEmailProviderTestSuite) TestGenerateBillingAndSentEmailProvider_ShouldRemove() {
	bankSlips := &bsEntities.BankSlipMap{
		"debit1": &bsEntities.BankSlip{
			DebtId: "debit1",
		},
		"debit2": &bsEntities.BankSlip{
			DebtId: "debit2",
		},
	}

	s.mockBillingService.On("GenerateBiling", mock.Anything).
		Return(&map[bsEntities.DebitId]error{"debit1": assert.AnError}).
		Once()
	s.mockEmailService.On("SendBankSlipWaitingPaymentEmail", mock.Anything).
		Return(&map[bsEntities.DebitId]error{"debit2": assert.AnError}).
		Once()

	rowsWithError := s.provider.GenerateBillingAndSentEmail(bankSlips)

	s.mockBillingService.AssertCalled(s.T(), "GenerateBiling", mock.MatchedBy(func(bankSlips *bsEntities.BankSlipMap) bool {
		return (*bankSlips)["debit1"].DebtId == "debit1" && (*bankSlips)["debit2"].DebtId == "debit2"
	}))
	s.mockEmailService.AssertCalled(s.T(), "SendBankSlipWaitingPaymentEmail", mock.MatchedBy(func(bankSlips *bsEntities.BankSlipMap) bool {
		return len(*bankSlips) == 1 && (*bankSlips)["debit2"].DebtId == "debit2"
	}))
	assert.Equal(s.T(), len(*bankSlips), 0)
	assert.Equal(s.T(), len(*rowsWithError), 2)
	assert.Equal(s.T(), bsEntities.BankSlipStatusGenerateBillingError, (*rowsWithError)["debit1"].Status)
	assert.Equal(s.T(), bsEntities.BankSlipStatusSendingEmailError, (*rowsWithError)["debit2"].Status)
}

func (s *GenerateBillingAndSentEmailProviderTestSuite) TestGenerateBillingAndSentEmailProvider_ShouldDoAllCorrectlyOnSuccess() {
	bankSlips := &bsEntities.BankSlipMap{
		"debit1": &bsEntities.BankSlip{
			DebtId: "debit1",
		},
		"debit2": &bsEntities.BankSlip{
			DebtId: "debit2",
		},
	}

	s.mockBillingService.On("GenerateBiling", mock.Anything).
		Return(&map[bsEntities.DebitId]error{}).
		Once()
	s.mockEmailService.On("SendBankSlipWaitingPaymentEmail", mock.Anything).
		Return(&map[bsEntities.DebitId]error{}).
		Once()

	s.provider.GenerateBillingAndSentEmail(bankSlips)

	s.mockBillingService.AssertCalled(s.T(), "GenerateBiling", mock.MatchedBy(func(bankSlips *bsEntities.BankSlipMap) bool {
		return (*bankSlips)["debit1"].DebtId == "debit1" && (*bankSlips)["debit2"].DebtId == "debit2"
	}))
	s.mockEmailService.AssertCalled(s.T(), "SendBankSlipWaitingPaymentEmail", mock.MatchedBy(func(bankSlips *bsEntities.BankSlipMap) bool {
		return (*bankSlips)["debit1"].DebtId == "debit1" && (*bankSlips)["debit2"].DebtId == "debit2"
	}))
	assert.Equal(s.T(), bsEntities.BankSlipStatusSuccess, (*bankSlips)["debit2"].Status)
	assert.Equal(s.T(), bsEntities.BankSlipStatusSuccess, (*bankSlips)["debit1"].Status)
}
