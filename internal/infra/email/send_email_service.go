package email

import bankSlipEntities "performatic-file-processor/internal/bank_slip/entity"

type EmailTemplate = string

var (
	BILLING_WAITING_PAYMENT EmailTemplate = "billing_waiting_payment"
)

type SendEmailService interface {
	SendBankSlipWaitingPaymentEmail(
		data *map[bankSlipEntities.DebitId]*bankSlipEntities.BankSlip,
	) *map[bankSlipEntities.DebitId]error
}
