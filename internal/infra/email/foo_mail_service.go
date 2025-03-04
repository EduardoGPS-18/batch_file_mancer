package mail

import bankSlipEntities "performatic-file-processor/internal/bank_slip/entity"

type FooSendMail struct {
}

type SentEmailData struct {
	To       string
	Subject  string
	Body     string
	DueDate  string
	Customer string
}

func NewFooSendMailService() *FooSendMail {
	return &FooSendMail{}
}

func (s *FooSendMail) sendMail(_ map[bankSlipEntities.DebitId]SentEmailData, _ []EmailTemplate) error {

	return nil
}

func (s *FooSendMail) SendBankSlipWaitingPaymentEmail(
	data []*bankSlipEntities.BankSlip,
) *map[bankSlipEntities.DebitId]error {
	toApi := map[bankSlipEntities.DebitId]SentEmailData{}
	for _, entity := range data {
		toApi[entity.DebtId] = SentEmailData{
			To:       entity.UserEmail,
			Subject:  "Billing Waiting Payment",
			Body:     "Your billing is waiting for payment",
			DueDate:  entity.DebtDueDate.String(),
			Customer: entity.UserName,
		}
	}

	// communicate to the email api
	// with the data in toApi

	s.sendMail(toApi, []EmailTemplate{BILLING_WAITING_PAYMENT})

	return &map[bankSlipEntities.DebitId]error{}
}
