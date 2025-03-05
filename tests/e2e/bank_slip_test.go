package e2e

import (
	"bytes"
	"context"
	"io"
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

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func TestUploadBankSlipFileHandler(t *testing.T) {

	containerFactory := sharedTestHelpers.NewContainerFactory(t.Context())

	kafkaContainer := containerFactory.MakeKafkaLandoopContainer()
	dbContainer := containerFactory.MakeDBContainer()

	defer kafkaContainer.Terminate(t.Context())
	defer dbContainer.Terminate(t.Context())

	dbInstance := database.GetInstance()

	router := httprouter.New()
	bankSlipRoutes.RegisterRoutes(router)

	consumer := bankSlipRoutes.NewBankSlipFactory().MakeBankSlipRowsConsumer(5)

	consumerCtx, cancel := context.WithCancel(context.Background())
	go consumer.Execute(consumerCtx, make(chan messaging.Message))

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	file, err := os.Open("../data/test_file.csv")
	if err != nil {
		t.Fatal(err)
	}

	assert.NoError(t, err)
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	assert.NoError(t, err)

	_, err = io.Copy(part, file)
	assert.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload/bank-slip/file", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	retries := 0
	for {
		time.Sleep(3 * time.Second)
		if retries == 10 {
			t.Fatal("Timeout waiting for bank slip processing")
		}
		queryRes, err := dbInstance.Query("select user_name, government_id, user_email, debt_amount, debt_due_date, debt_id from bank_slip;")
		if err != nil {
			t.Fatal(err)
		}

		found := false
		for queryRes.Next() {
			var name, governmentId, email, debtId string
			var debtAmount float64
			var debtDueDate time.Time
			err = queryRes.Scan(&name, &governmentId, &email, &debtAmount, &debtDueDate, &debtId)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "Elijah Santos", name)
			assert.Equal(t, "9558", governmentId)
			assert.Equal(t, "janet95@example.com", email)
			assert.Equal(t, 7811.0, debtAmount)
			assert.Equal(t, time.Date(2024, 1, 19, 0, 0, 0, 0, time.UTC), debtDueDate)
			assert.Equal(t, "ea23f2ca-663a-4266-a742-9da4c9f4fcb3", debtId)
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
