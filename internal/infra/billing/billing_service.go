package billing

import bankSlipEntities "performatic-file-processor/internal/bank_slip/entity"

type BilingService interface {
	GenerateBiling(bankSlip *map[bankSlipEntities.DebitId]*bankSlipEntities.BankSlip) *map[bankSlipEntities.DebitId]error
}
