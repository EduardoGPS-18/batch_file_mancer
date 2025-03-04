package bank_slip

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	bankSlipEntities "performatic-file-processor/internal/bank_slip/entity"
	entities "performatic-file-processor/internal/bank_slip/entity"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuitBankSlipPgRepository struct {
	suite.Suite
	db         *sql.DB
	mock       sqlmock.Sqlmock
	repository *BankSlipPgRepository
}

func (testSuit *TestSuitBankSlipPgRepository) SetupTest() {
	db, mock, err := sqlmock.New()
	assert.NoError(testSuit.T(), err)
	testSuit.db = db
	testSuit.mock = mock
	testSuit.repository = NewBankSlipPgRepository(db)
}

func TestBankSlipPgRepository(t *testing.T) {
	suite.Run(t, new(TestSuitBankSlipPgRepository))
}

func (s *TestSuitBankSlipPgRepository) TestBankSlipPgRepository_InsertMany() {
	bankSlips := map[bankSlipEntities.DebitId]*bankSlipEntities.BankSlip{
		"1": {
			UserName:               "John Doe",
			GovernmentId:           5321,
			UserEmail:              "johndoe@example.com",
			DebtAmount:             1000.00,
			DebtDueDate:            time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
			DebtId:                 "1",
			BankSlipFileMetadataId: "file_123",
			Status:                 "pending",
			ErrorMessage:           nil,
		},
	}

	s.mock.ExpectExec("INSERT INTO bank_slip").
		WithArgs(
			"John Doe", 5321, "johndoe@example.com", 1000.00, time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), "1", "file_123", "pending", nil,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := s.repository.InsertMany(bankSlips)
	assert.NoError(s.T(), err)
}

func (s *TestSuitBankSlipPgRepository) TestBankSlipPgRepository_InsertMany_LastWithoutComa() {
	errorMsg := "any_error"
	bankSlips := map[entities.DebitId]*entities.BankSlip{
		"1": &bankSlipEntities.BankSlip{
			UserName:               "John Doe",
			GovernmentId:           5421,
			UserEmail:              "john.doe@example.com",
			DebtAmount:             1000.50,
			DebtDueDate:            time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
			DebtId:                 "1",
			BankSlipFileMetadataId: "file1",
			Status:                 "pending",
			ErrorMessage:           nil,
		},
		"2": &bankSlipEntities.BankSlip{
			UserName:               "Jane Doe",
			GovernmentId:           7632,
			UserEmail:              "jane.doe@example.com",
			DebtAmount:             2000.75,
			DebtDueDate:            time.Date(2025, 8, 31, 0, 0, 0, 0, time.UTC),
			DebtId:                 "2",
			BankSlipFileMetadataId: "file2",
			Status:                 "paid",
			ErrorMessage:           &errorMsg,
		},
	}

	// Configura a expectativa para a query no mock do banco de dados
	s.mock.ExpectExec("INSERT INTO bank_slip").WithArgs(
		"John Doe", 5421, "john.doe@example.com", 1000.50, time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), "1", "file1", "pending", nil,
		"Jane Doe", 7632, "jane.doe@example.com", 2000.75, time.Date(2025, 8, 31, 0, 0, 0, 0, time.UTC), "2", "file2", "paid", &errorMsg,
	).WillReturnResult(sqlmock.NewResult(1, 2))

	// Chama o método InsertMany
	err := s.repository.InsertMany(bankSlips)
	assert.NoError(s.T(), err)

	// Verifica se as expectativas do mock foram atendidas
	err = s.mock.ExpectationsWereMet()
	assert.NoError(s.T(), err)
}

func (s *TestSuitBankSlipPgRepository) TestBankSlipPgRepository_InsertMany_Error() {
	bankSlips := map[bankSlipEntities.DebitId]*bankSlipEntities.BankSlip{
		"1": {
			UserName:               "John Doe",
			GovernmentId:           5321,
			UserEmail:              "johndoe@example.com",
			DebtAmount:             1000.00,
			DebtDueDate:            time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
			DebtId:                 "1",
			BankSlipFileMetadataId: "file_123",
			Status:                 "pending",
			ErrorMessage:           nil,
		},
	}

	s.mock.ExpectExec("INSERT INTO bank_slip").
		WithArgs(
			"John Doe", 5321, "johndoe@example.com", 1000.00, time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), "1", "file_123", "pending", nil,
		).
		WillReturnError(fmt.Errorf("insert error"))

	err := s.repository.InsertMany(bankSlips)
	assert.Error(s.T(), err)
	assert.EqualError(s.T(), err, "insert error")

	err = s.mock.ExpectationsWereMet()
	assert.NoError(s.T(), err)
}

func (s *TestSuitBankSlipPgRepository) TestBankSlipPgRepository_GetExistingByDebitIds() {
	debitIds := []string{"1", "2", "3"}
	rows := sqlmock.NewRows([]string{"debt_id"}).
		AddRow("1").
		AddRow("2")

	s.mock.ExpectQuery("SELECT debt_id FROM bank_slip WHERE debt_id IN").
		WithArgs("1", "2", "3").
		WillReturnRows(rows)

	result, err := s.repository.GetExistingByDebitIds(debitIds)
	assert.NoError(s.T(), err)

	assert.Len(s.T(), result, 2)
	assert.Contains(s.T(), result, "1")
	assert.Contains(s.T(), result, "2")
	assert.NotContains(s.T(), result, "3")

	err = s.mock.ExpectationsWereMet()
	assert.NoError(s.T(), err)
}

func (s *TestSuitBankSlipPgRepository) TestBankSlipPgRepository_GetExistingByDebitIds_Error() {
	debitIds := []string{"1", "2", "3"}

	s.mock.ExpectQuery("SELECT debt_id FROM bank_slip WHERE debt_id IN").
		WithArgs("1", "2", "3").
		WillReturnError(fmt.Errorf("query error"))

	result, err := s.repository.GetExistingByDebitIds(debitIds)
	assert.Error(s.T(), err)
	assert.Nil(s.T(), result)

	err = s.mock.ExpectationsWereMet()
	assert.NoError(s.T(), err)
}
