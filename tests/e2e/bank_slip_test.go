package e2e

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	bankSlipRoutes "performatic-file-processor/internal/bank_slip/routes"
	"performatic-file-processor/internal/database"
	"performatic-file-processor/internal/messaging"
	sharedTestHelpers "performatic-file-processor/tests/shared"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

type BankSlipTestE2ESuite struct {
	suite.Suite
	kafkaContainer testcontainers.Container
	dbContainer    testcontainers.Container
	dbInstance     *sql.DB
}

func TestBankSlipE2eRunSuite(t *testing.T) {
	suite.Run(t, new(BankSlipTestE2ESuite))
}

func (f *BankSlipTestE2ESuite) SetupSuite() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	containerFactory := sharedTestHelpers.NewContainerFactory(f.T().Context())
	dbInstance := database.GetInstance()

	f.dbContainer = containerFactory.MakeDBContainer()
	f.kafkaContainer = containerFactory.MakeKafkaLandoopContainer()
	f.dbInstance = dbInstance
}

func (f *BankSlipTestE2ESuite) TearDownTest() {
	defer f.kafkaContainer.Terminate(f.T().Context())
	defer f.dbContainer.Terminate(f.T().Context())
}

func (f *BankSlipTestE2ESuite) TestBankSlipE2eRunSuite_UploadFileAndProcessRows() {
	router := httprouter.New()
	bankSlipRoutes.RegisterRoutes(router)

	consumer := bankSlipRoutes.NewBankSlipFactory().MakeBankSlipRowsConsumer(5)

	consumerCtx, cancel := context.WithCancel(context.Background())
	go consumer.Execute(consumerCtx, make(chan messaging.Message))

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	file, err := os.Open("./data/test_file.csv")
	if err != nil {
		f.T().Fatal(err)
	}

	assert.NoError(f.T(), err)
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	assert.NoError(f.T(), err)

	_, err = io.Copy(part, file)
	assert.NoError(f.T(), err)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload/bank-slip/file", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(f.T(), http.StatusOK, rr.Code)

	retries := 0
	for {
		time.Sleep(3 * time.Second)
		if retries == 10 {
			f.T().Fatal("Timeout waiting for bank slip processing")
		}
		queryRes, err := f.dbInstance.Query("select user_name, government_id, user_email, debt_amount, debt_due_date, debt_id from bank_slip;")
		if err != nil {
			f.T().Fatal(err)
		}

		found := false
		for queryRes.Next() {
			var name, governmentId, email, debtId string
			var debtAmount float64
			var debtDueDate time.Time
			err = queryRes.Scan(&name, &governmentId, &email, &debtAmount, &debtDueDate, &debtId)
			if err != nil {
				f.T().Fatal(err)
			}
			assert.Equal(f.T(), "Elijah Santos", name)
			assert.Equal(f.T(), "9558", governmentId)
			assert.Equal(f.T(), "janet95@example.com", email)
			assert.Equal(f.T(), 7811.0, debtAmount)
			assert.Equal(f.T(), time.Date(2024, 1, 19, 0, 0, 0, 0, time.UTC), debtDueDate)
			assert.Equal(f.T(), "ea23f2ca-663a-4266-a742-9da4c9f4fcb3", debtId)
			found = true
		}
		if found {
			cancel()
			time.Sleep(3 * time.Second)
			break
		}

		retries++
	}
}
