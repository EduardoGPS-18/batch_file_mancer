package billing

import (
	bankSlipEntities "performatic-file-processor/internal/bank_slip/entity"
)

type FooBillingService struct {
}

type GenerateBillingData struct {
	Amount   float64
	DueDate  string
	Customer string
}

func NewFooBillingService() *FooBillingService {
	return &FooBillingService{}
}

func (f FooBillingService) GenerateBiling(
	bankSlips *map[bankSlipEntities.DebitId]*bankSlipEntities.BankSlip,
) *map[bankSlipEntities.DebitId]error {
	toApi := map[string]any{}

	for _, entity := range *bankSlips {
		toApi[entity.DebtId] = GenerateBillingData{
			Amount:   entity.DebtAmount,
			DueDate:  entity.DebtDueDate.Format("2006-01-02"),
			Customer: entity.UserEmail,
		}
	}

	// communicate to the billing api
	// with the data in toApi

	return &map[bankSlipEntities.DebitId]error{}
}
