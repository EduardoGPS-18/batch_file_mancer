package integration

import (
	"database/sql"
	"log"
	bankSlipEntities "performatic-file-processor/internal/bank_slip/entity"
	bankSlipRepositories "performatic-file-processor/internal/bank_slip/repositories"
	"performatic-file-processor/internal/database"
	"testing"
	"time"

	sharedTestHelpers "performatic-file-processor/tests/shared"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type BankSlipTestIntegration struct {
	suite.Suite
	bankSlipRepository     *bankSlipRepositories.BankSlipPgRepository
	bankSlipFileRepository *bankSlipRepositories.BankSlipFilePgRepository
	postgresContainer      testcontainers.Container
	db                     *sql.DB
}

func (s *BankSlipTestIntegration) SetupSuite() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	log.Print("\nSetting up test suite...\n")

	containerFactory := sharedTestHelpers.NewContainerFactory(s.T().Context())

	s.postgresContainer = containerFactory.MakeDBContainer()
}

func (s *BankSlipTestIntegration) SetupTest() {
	db := database.GetInstance()
	s.bankSlipRepository = bankSlipRepositories.NewBankSlipPgRepository(db)
	s.bankSlipFileRepository = bankSlipRepositories.NewBankSlipFilePgRepository(db)
	s.db = db
}

func (s *BankSlipTestIntegration) TearDownSuite() {
	defer s.postgresContainer.Terminate(s.T().Context())
}

func TestBankSlipRunSuite(t *testing.T) {
	suite.Run(t, new(BankSlipTestIntegration))
}

func (f *BankSlipTestIntegration) TestBankSlipTest_ShouldInsertBankSlipFileCorrectly() {
	f.bankSlipFileRepository.Insert(&bankSlipEntities.BankSlipFileMetadata{FileName: "test.csv"})

	queryResult, err := f.db.Query(`select id, name, created_at from bank_slip_file where name = 'test.csv'`)
	if err != nil {
		log.Fatalf("Error querying bank_slip_file: %v", err)
	}

	var id, fileName string
	var createdAt time.Time

	for queryResult.Next() {
		err = queryResult.Scan(&id, &fileName, &createdAt)
	}

	assert.NoError(f.T(), err)

	assert.NotZero(f.T(), id)
	assert.Equal(f.T(), "test.csv", fileName)
	assert.NotZero(f.T(), createdAt)
}

func (f *BankSlipTestIntegration) TestBankSlipTest_ShouldInsertBankSlipCorrectly() {
	queryResult, err := f.db.Query(`INSERT INTO bank_slip_file (name) VALUES ('test.csv') RETURNING id`)
	if err != nil {
		log.Fatalf("Error inserting into bank_slip_file: %v", err)
	}

	var fileId string
	for queryResult.Next() {
		queryResult.Scan(&fileId)
	}

	debtId := uuid.New().String()
	f.bankSlipRepository.InsertMany(&bankSlipEntities.BankSlipMap{
		"test.csv": &bankSlipEntities.BankSlip{
			UserName:               "Test User",
			DebtId:                 debtId,
			DebtAmount:             100.51,
			DebtDueDate:            time.Now(),
			GovernmentId:           521,
			UserEmail:              "test@user.com",
			BankSlipFileMetadataId: fileId,
			Status:                 "PENDING",
			ErrorMessage:           nil,
		},
	})
	query := `SELECT debt_id, debt_amount, debt_due_date, user_name, government_id, user_email, 
                     bank_slip_file_id, status, error_message FROM bank_slip`
	rows, err := f.db.Query(query)
	if err != nil {
		log.Fatalf("Error querying bank_slip: %v", err)
	}
	var bankSlip bankSlipEntities.BankSlip
	for rows.Next() {
		err := rows.Scan(&bankSlip.DebtId, &bankSlip.DebtAmount, &bankSlip.DebtDueDate, &bankSlip.UserName,
			&bankSlip.GovernmentId, &bankSlip.UserEmail, &bankSlip.BankSlipFileMetadataId,
			&bankSlip.Status, &bankSlip.ErrorMessage)
		if err != nil {
			panic(err)
		}
	}
	assert.NotZero(f.T(), bankSlip.DebtId, debtId)
	assert.Equal(f.T(), bankSlip.DebtAmount, 100.51)
	assert.NotZero(f.T(), bankSlip.DebtDueDate)
	assert.Equal(f.T(), bankSlip.UserName, "Test User")
	assert.Equal(f.T(), bankSlip.GovernmentId, 521)
	assert.Equal(f.T(), bankSlip.UserEmail, "test@user.com")
	assert.Equal(f.T(), bankSlip.BankSlipFileMetadataId, fileId)
	assert.Equal(f.T(), bankSlip.Status, bankSlipEntities.BankSlipStatusPending)
	assert.Nil(f.T(), bankSlip.ErrorMessage)
}

func (f *BankSlipTestIntegration) TestBankSlipTest_ShouldUpdateBankSlipCorrectly() {
	queryResult, err := f.db.Query(`INSERT INTO bank_slip_file (name) VALUES ('test.csv') RETURNING id`)
	if err != nil {
		log.Fatalf("Error inserting into bank_slip_file: %v", err)
	}

	var fileId string
	for queryResult.Next() {
		queryResult.Scan(&fileId)
	}

	debtId := uuid.New().String()
	bankSlip := &bankSlipEntities.BankSlip{
		UserName:               "Test User",
		DebtId:                 debtId,
		DebtAmount:             100.51,
		DebtDueDate:            time.Now(),
		GovernmentId:           521,
		UserEmail:              "test@user.com",
		BankSlipFileMetadataId: fileId,
		Status:                 "PENDING",
		ErrorMessage:           nil,
	}
	f.bankSlipRepository.InsertMany(&bankSlipEntities.BankSlipMap{
		"test.csv": bankSlip,
	})

	// Update the bank slip
	bankSlip.Status = bankSlipEntities.BankSlipStatusSuccess
	bankSlip.DebtAmount = 0
	err = f.bankSlipRepository.UpdateMany(&bankSlipEntities.BankSlipMap{
		"any_key": bankSlip,
	})
	if err != nil {
		log.Fatalf("Error updating bank_slip: %v", err)
	}

	// Verify the update
	query := `SELECT debt_id, debt_amount, debt_due_date, user_name, government_id, user_email, 
					 bank_slip_file_id, status, error_message FROM bank_slip WHERE debt_id = $1`
	row := f.db.QueryRow(query, debtId)
	var updatedBankSlip bankSlipEntities.BankSlip
	err = row.Scan(
		&updatedBankSlip.DebtId,
		&updatedBankSlip.DebtAmount,
		&updatedBankSlip.DebtDueDate,
		&updatedBankSlip.UserName,
		&updatedBankSlip.GovernmentId,
		&updatedBankSlip.UserEmail,
		&updatedBankSlip.BankSlipFileMetadataId,
		&updatedBankSlip.Status,
		&updatedBankSlip.ErrorMessage,
	)
	if err != nil {
		log.Fatalf("Error querying updated bank_slip: %v", err)
	}

	assert.Equal(f.T(), updatedBankSlip.DebtId, debtId)
	assert.Equal(f.T(), updatedBankSlip.DebtAmount, 100.51)
	assert.NotZero(f.T(), updatedBankSlip.DebtDueDate)
	assert.Equal(f.T(), updatedBankSlip.UserName, "Test User")
	assert.Equal(f.T(), updatedBankSlip.GovernmentId, 521)
	assert.Equal(f.T(), updatedBankSlip.UserEmail, "test@user.com")
	assert.Equal(f.T(), updatedBankSlip.BankSlipFileMetadataId, fileId)
	assert.Equal(f.T(), updatedBankSlip.Status, bankSlipEntities.BankSlipStatusSuccess)
	assert.Nil(f.T(), updatedBankSlip.ErrorMessage)
}
