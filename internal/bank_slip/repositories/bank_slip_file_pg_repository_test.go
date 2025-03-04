package bank_slip

import (
	"database/sql"
	bankSlipEntities "performatic-file-processor/internal/bank_slip/entity"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type BankSlipFilePgRepositoryTestSuite struct {
	suite.Suite
	repository *BankSlipFilePgRepository
	db         *sql.DB
	mock       sqlmock.Sqlmock
}

func (testSuit *BankSlipFilePgRepositoryTestSuite) SetupTest() {
	db, mock, err := sqlmock.New()
	assert.NoError(testSuit.T(), err)
	testSuit.db = db
	testSuit.mock = mock
	testSuit.repository = NewBankSlipFilePgRepository(db)
}

func TestBankSlipFilePgRepository(t *testing.T) {
	suite.Run(t, new(BankSlipFilePgRepositoryTestSuite))
}

func (suite *BankSlipFilePgRepositoryTestSuite) TestInsertAndGetBankSlipFileMetadata() {
	fileMetadata := &bankSlipEntities.BankSlipFileMetadata{
		FileName: "test_file.txt",
	}

	suite.mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO bank_slip_file (name) VALUES ($1) returning id")).
		WithArgs(fileMetadata.FileName).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := suite.repository.Insert(fileMetadata)
	assert.NoError(suite.T(), err)

	err = suite.mock.ExpectationsWereMet()
	assert.NoError(suite.T(), err) // Ensure no unmatched expectations

}

func (suite *BankSlipFilePgRepositoryTestSuite) TestInsertAndGetBankSlipFileMetadataError() {
	fileMetadata := &bankSlipEntities.BankSlipFileMetadata{
		FileName: "test_file.txt",
	}

	suite.mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO bank_slip_file (name) VALUES ($1) returning id")).
		WithArgs(fileMetadata.FileName).
		WillReturnError(sql.ErrNoRows)

	err := suite.repository.Insert(fileMetadata)
	assert.Error(suite.T(), err)

}
