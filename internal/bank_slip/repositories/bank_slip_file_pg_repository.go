package repositories

import (
	"database/sql"
	"errors"

	entities "performatic-file-processor/internal/bank_slip/entity"
)

type BankSlipFilePgRepository struct {
	db *sql.DB
}

func NewBankSlipFilePgRepository(db *sql.DB) *BankSlipFilePgRepository {
	return &BankSlipFilePgRepository{db: db}
}

func (r *BankSlipFilePgRepository) Insert(bankSlipFile *entities.BankSlipFileMetadata) error {
	query := "INSERT INTO bank_slip_file (name) VALUES ($1) returning id"

	err := r.db.QueryRow(query, bankSlipFile).Scan(&bankSlipFile.ID)

	if err != nil {
		return errors.New("erro ao inserir arquivo no banco")
	}
	return nil
}
