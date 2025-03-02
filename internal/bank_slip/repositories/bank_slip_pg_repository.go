package bank_slip

import (
	"database/sql"
	"fmt"
	"strings"

	entities "performatic-file-processor/internal/bank_slip/entity"
)

type BankSlipPgRepository struct {
	db *sql.DB
}

func NewBankSlipPgRepository(db *sql.DB) *BankSlipPgRepository {
	return &BankSlipPgRepository{db: db}
}

func (r *BankSlipPgRepository) InsertMany(bankSlips map[entities.DebitId]*entities.BankSlip) error {

	fields := []any{}
	queryValues := ""
	i := 0
	for _, slip := range bankSlips {
		fields = append(fields, slip.UserName, slip.GovernmentId, slip.UserEmail, slip.DebtAmount, slip.DebtDueDate, slip.DebtId, slip.BankSlipFileMetadataId, slip.Status, slip.ErrorMessage)
		queryValues += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*9+1, i*9+2, i*9+3, i*9+4, i*9+5, i*9+6, i*9+7, i*9+8, i*9+9)
		if i < len(bankSlips)-1 {
			queryValues += ", "
		}
		i++
	}

	query := fmt.Sprintf("INSERT INTO bank_slip (user_name, government_id, user_email, debt_amount, debt_due_date, debt_id, bank_slip_file_id, status, error_message) VALUES %s", queryValues)
	_, err := r.db.Exec(query, fields...)
	if err != nil {
		return err
	}
	return nil
}

func (r *BankSlipPgRepository) GetExistingByDebitIds(debitIds []string) (map[entities.DebitId]entities.Existing, error) {
	joinedDebitIds := strings.Join(debitIds, ",")
	selectQuery := fmt.Sprintf("SELECT debt_id FROM bank_slip WHERE debt_id IN (%s)", joinedDebitIds)
	rows, err := r.db.Query(selectQuery)
	if err != nil {
		return nil, err
	}

	bankSlips := map[entities.DebitId]entities.Existing{}
	for rows.Next() {
		var debtId string
		rows.Scan(&debtId)
		bankSlips[debtId] = true
	}
	return bankSlips, nil
}
