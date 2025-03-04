package billing

import (
	bankSlipEntities "performatic-file-processor/internal/bank_slip/entity"
)

type FooBillingService struct {
}

func (f FooBillingService) GenerateBiling(entities []*bankSlipEntities.BankSlip) *map[bankSlipEntities.DebitId]error {
	toApi := map[string]any{}
	for _, entity := range entities {
		toApi[entity.DebtId] = map[string]any{
			"amount":   entity.DebtAmount,
			"due_date": entity.DebtDueDate,
			"customer": entity.UserEmail,
		}
	}

	// communicate to the billing api
	// with the data in toApi

	return &map[bankSlipEntities.DebitId]error{}
}
