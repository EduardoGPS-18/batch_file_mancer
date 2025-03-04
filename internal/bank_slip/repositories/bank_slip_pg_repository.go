package bank_slip

import (
	"database/sql"
	"fmt"
	"log"

	entities "performatic-file-processor/internal/bank_slip/entity"
)

type BankSlipPgRepository struct {
	db *sql.DB
}

func NewBankSlipPgRepository(db *sql.DB) *BankSlipPgRepository {
	return &BankSlipPgRepository{db: db}
}

func (r *BankSlipPgRepository) UpdateMany(bankSlipList ...*entities.BankSlipMap) error {
	fields := []any{}
	queryValues := ""
	i := 0
	for _, bankSlipP := range bankSlipList {
		bankSlip := *bankSlipP
		for _, slip := range bankSlip {
			fields = append(fields, slip.DebtId, slip.Status, slip.ErrorMessage)
			queryValues += fmt.Sprintf("(cast($%d AS uuid), $%d, $%d)", i*3+1, i*3+2, i*3+3)
			if i < len(bankSlip)-1 {
				queryValues += ", "
			}
			i++
		}
	}
	query := fmt.Sprintf(`
		UPDATE bank_slip bs 
		SET
			status = tmp.status,
			error_message = tmp.error_message
		FROM (
			VALUES
				%s
		) AS tmp(debt_id, status, error_message)
		WHERE bs.debt_id = tmp.debt_id
	`, queryValues)
	_, err := r.db.Exec(query, fields...)
	if err != nil {
		return err
	}
	return nil
}

func (r *BankSlipPgRepository) InsertMany(bankSlipsP *entities.BankSlipMap) (map[entities.DebitId]entities.Success, error) {
	bankSlips := *bankSlipsP
	fields := []any{}
	queryValues := ""
	i := 0
	insertedDebtIds := map[entities.DebitId]entities.Success{}
	for _, slip := range bankSlips {
		insertedDebtIds[slip.DebtId] = false
		fields = append(fields, slip.UserName, slip.GovernmentId, slip.UserEmail, slip.DebtAmount, slip.DebtDueDate, slip.DebtId, slip.BankSlipFileMetadataId, slip.Status, slip.ErrorMessage)
		queryValues += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", i*9+1, i*9+2, i*9+3, i*9+4, i*9+5, i*9+6, i*9+7, i*9+8, i*9+9)
		if i < len(bankSlips)-1 {
			queryValues += ", "
		}
		i++
	}

	query := fmt.Sprintf("INSERT INTO bank_slip (user_name, government_id, user_email, debt_amount, debt_due_date, debt_id, bank_slip_file_id, status, error_message) VALUES %s ON CONFLICT DO NOTHING RETURNING debt_id", queryValues)
	queryResult, err := r.db.Query(query, fields...)
	if err != nil {
		return nil, err
	}

	for queryResult.Next() {
		var debtId string
		if err := queryResult.Scan(&debtId); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		insertedDebtIds[debtId] = true
	}

	return insertedDebtIds, nil
}
