package bank_slip

import (
	bsEntities "performatic-file-processor/internal/bank_slip/entity"
	"performatic-file-processor/internal/infra/billing"
	emailService "performatic-file-processor/internal/infra/email"
)

type GenerateBillingAndSentEmailProvider interface {
	GenerateBillingAndSentEmail(
		bankSlips *bsEntities.BankSlipMap,
	) *bsEntities.BankSlipMap
}

type GenerateBillingAndSentEmailProviderImpl struct {
	emailService   emailService.EmailService
	billingService billing.BilingService
}

func NewGenerateBillingAndSentEmailProvider(
	emailService emailService.EmailService,
	billingService billing.BilingService,
) *GenerateBillingAndSentEmailProviderImpl {
	return &GenerateBillingAndSentEmailProviderImpl{
		emailService:   emailService,
		billingService: billingService,
	}
}

func (p *GenerateBillingAndSentEmailProviderImpl) GenerateBillingAndSentEmail(
	bankSlips *bsEntities.BankSlipMap,
) *bsEntities.BankSlipMap {
	successBankSlips := *bankSlips
	bankSlipsWithError := map[bsEntities.DebitId]*bsEntities.BankSlip{}

	errorsGeneratingBilling := *p.billingService.GenerateBiling(&successBankSlips)
	for debtId := range errorsGeneratingBilling {
		bankSlipWithError := successBankSlips[debtId]
		bankSlipWithError.ErrorGeneratingBilling(errorsGeneratingBilling[debtId].Error())
		if bankSlipWithError != nil {
			bankSlipsWithError[debtId] = bankSlipWithError
			delete(successBankSlips, debtId)
		}
	}

	errorsSendingEmail := *p.emailService.SendBankSlipWaitingPaymentEmail(&successBankSlips)
	for debtId := range errorsSendingEmail {
		bankSlipWithError := successBankSlips[debtId]
		bankSlipWithError.ErrorSendingEmail(errorsSendingEmail[debtId].Error())
		if bankSlipWithError != nil {
			bankSlipsWithError[debtId] = bankSlipWithError
			delete(successBankSlips, debtId)
		}
	}

	for _, bankSlip := range successBankSlips {
		bankSlip.Success()
	}
	return &bankSlipsWithError
}
