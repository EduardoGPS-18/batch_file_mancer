package bank_slip

import (
	bankSlipEntities "performatic-file-processor/internal/bank_slip/entity"
	"performatic-file-processor/internal/infra/billing"
	emailService "performatic-file-processor/internal/infra/email"
)

type GenerateBillingAndSentEmailResponse struct {
	ErrorGeneratingBilling *map[bankSlipEntities.DebitId]*bankSlipEntities.BankSlip
	ErrorSendingEmail      *map[bankSlipEntities.DebitId]*bankSlipEntities.BankSlip
	Success                *map[bankSlipEntities.DebitId]*bankSlipEntities.BankSlip
}

type GenerateBillingAndSentEmailProvider interface {
	GenerateBillingAndSentEmail(
		bankSlips *map[bankSlipEntities.DebitId]*bankSlipEntities.BankSlip,
	) GenerateBillingAndSentEmailResponse
}

type GenerateBillingAndSentEmailProviderImpl struct {
	emailService   emailService.SendEmailService
	billingService billing.BilingService
}

func NewGenerateBillingAndSentEmailProvider(
	emailService emailService.SendEmailService,
	billingService billing.BilingService,
) *GenerateBillingAndSentEmailProviderImpl {
	return &GenerateBillingAndSentEmailProviderImpl{
		emailService:   emailService,
		billingService: billingService,
	}
}

func (p *GenerateBillingAndSentEmailProviderImpl) GenerateBillingAndSentEmail(
	bankSlips *map[bankSlipEntities.DebitId]*bankSlipEntities.BankSlip,
) GenerateBillingAndSentEmailResponse {

}
