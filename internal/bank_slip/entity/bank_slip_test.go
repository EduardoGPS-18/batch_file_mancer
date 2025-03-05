package bank_slip

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewBankSlip(t *testing.T) {
	debtDueDate, _ := time.Parse("2006-01-02", "2023-12-31")
	bankSlip := newBankSlip(123, 1000.50, debtDueDate, "debt123", "John Doe", "john.doe@example.com", "file123", BankSlipStatusPending)

	assert.Equal(t, 123, bankSlip.GovernmentId)
	assert.Equal(t, 1000.50, bankSlip.DebtAmount)
	assert.Equal(t, debtDueDate, bankSlip.DebtDueDate)
	assert.Equal(t, "debt123", bankSlip.DebtId)
	assert.Equal(t, "John Doe", bankSlip.UserName)
	assert.Equal(t, "john.doe@example.com", bankSlip.UserEmail)
	assert.Equal(t, "file123", bankSlip.BankSlipFileMetadataId)
	assert.Equal(t, BankSlipStatusPending, bankSlip.Status)
}

func TestNewBankSlipFromRow(t *testing.T) {
	header := "name,governmentId,email,debtAmount,debtDueDate,debtId"
	data := "John Doe,123,john.doe@example.com,1000.50,2023-12-31,debt123"
	fileMetadataId := "file123"

	bankSlip, err := NewBankSlipFromRow(fileMetadataId, data, header)
	assert.NoError(t, err)

	debtDueDate, _ := time.Parse("2006-01-02", "2023-12-31")
	assert.Equal(t, 123, bankSlip.GovernmentId)
	assert.Equal(t, 1000.50, bankSlip.DebtAmount)
	assert.Equal(t, debtDueDate, bankSlip.DebtDueDate)
	assert.Equal(t, "debt123", bankSlip.DebtId)
	assert.Equal(t, "John Doe", bankSlip.UserName)
	assert.Equal(t, "john.doe@example.com", bankSlip.UserEmail)
	assert.Equal(t, "file123", bankSlip.BankSlipFileMetadataId)
	assert.Equal(t, BankSlipStatusPending, bankSlip.Status)
}

func TestNewBankSlipFromRow_HeaderAndRowIsDiferent(t *testing.T) {
	header := "name,governmentId,email,debtAmount,debtDueDate,debtId"

	data := "John Doe,321,john.doe@example.com,1000.50,2023-12-31"
	fileMetadataId := ""

	bankSlip, err := NewBankSlipFromRow(fileMetadataId, data, header)
	assert.Error(t, err)
	assert.Nil(t, bankSlip)
	assert.Equal(t, errors.New("file metadata must be not empty"), err)
}

func TestNewBankSlipFromRow_FileMetadataIsEmpty(t *testing.T) {
	header := "name,governmentId,email,debtAmount,debtDueDate,debtId"

	data := "John Doe,321,john.doe@example.com,1000.50,2023-12-31"
	fileMetadataId := "file123"

	bankSlip, err := NewBankSlipFromRow(fileMetadataId, data, header)
	assert.Error(t, err)
	assert.Nil(t, bankSlip)
	assert.Equal(t, errors.New("error rowItems and headerItems length are different (file id: file123)"), err)
}

func TestNewBankSlipFromRow_RowIsEmpty(t *testing.T) {
	header := "name,email,debtAmount,debtDueDate,debtId"

	data := ""
	fileMetadataId := "file123"

	bankSlip, err := NewBankSlipFromRow(fileMetadataId, data, header)
	assert.Error(t, err)
	assert.Nil(t, bankSlip)
	assert.Equal(t, errors.New("error is missing some field (file id: file123)"), err)
}

func TestNewBankSlipFromRow_SomeFieldNotProvidedError(t *testing.T) {
	for _, field := range []string{"name", "governmentId", "email", "debtAmount", "debtDueDate", "debtId"} {
		header := "name,governmentId,email,debtAmount,debtDueDate,debtId"
		headerWithoutField := strings.Replace(header, field, "missing", -1)

		data := "John Doe,321,john.doe@example.com,1000.50,2023-12-31,debt123"
		fileMetadataId := "file123"

		bankSlip, err := NewBankSlipFromRow(fileMetadataId, data, headerWithoutField)
		assert.Error(t, err)
		assert.Nil(t, bankSlip)
		assert.Equal(t, errors.New("error is missing some field (file id: file123)"), err)
	}
}

func TestNewBankSlipFromRow_InvalidGovernmentIdError(t *testing.T) {
	header := "name,governmentId,email,debtAmount,debtDueDate,debtId"
	data := "John Doe,abc,john.doe@example.com,1000.50,2023-12-31,debt123"
	fileMetadataId := "file123"

	bankSlip, err := NewBankSlipFromRow(fileMetadataId, data, header)
	assert.Error(t, err)
	assert.Nil(t, bankSlip)
	assert.Equal(t, errors.New("error converting governmentId to int abc Position: 1 (file id: file123)"), err)
}

func TestNewBankSlipFromRow_DebitAmountError(t *testing.T) {
	header := "name,governmentId,email,debtAmount,debtDueDate,debtId"
	data := "John Doe,123,john.doe@example.com,cde,2023-12-31,debt123"
	fileMetadataId := "file123"

	bankSlip, err := NewBankSlipFromRow(fileMetadataId, data, header)
	assert.Error(t, err)
	assert.Nil(t, bankSlip)
	assert.Equal(t, errors.New("error converting debtAmount to float64 cde Position: 3 (file id: file123)"), err)
}

func TestNewBankSlipFromRow_DebitDueDateError(t *testing.T) {
	header := "name,governmentId,email,debtAmount,debtDueDate,debtId"
	data := "John Doe,123,john.doe@example.com,1500.5,2023-12-32,debt123"
	fileMetadataId := "file123"

	bankSlip, err := NewBankSlipFromRow(fileMetadataId, data, header)
	assert.Error(t, err)
	assert.Nil(t, bankSlip)
	assert.Equal(t, errors.New("error converting debtDueDate to time.Time 2023-12-32 Position: 4 (file id: file123)"), err)
}

func TestUpdateRowToErrorGeneratingBilling(t *testing.T) {
	debtDueDate, _ := time.Parse("2006-01-02", "2023-12-31")
	bankSlip := newBankSlip(123, 1000.50, debtDueDate, "debt123", "John Doe", "john.doe@example.com", "file123", BankSlipStatusPending)

	errorMessage := "Any error"
	bankSlip.ErrorGeneratingBilling(errorMessage)

	assert.Equal(t, BankSlipStatusGenerateBillingError, bankSlip.Status)
	assert.Equal(t, &errorMessage, bankSlip.ErrorMessage)
}

func TestUpdateRowToErrorSendingEmail(t *testing.T) {
	debtDueDate, _ := time.Parse("2006-01-02", "2023-12-31")
	bankSlip := newBankSlip(123, 1000.50, debtDueDate, "debt123", "John Doe", "john.doe@example.com", "file123", BankSlipStatusPending)

	errorMessage := "Any error"
	bankSlip.ErrorSendingEmail(errorMessage)

	assert.Equal(t, BankSlipStatusSendingEmailError, bankSlip.Status)
	assert.Equal(t, &errorMessage, bankSlip.ErrorMessage)
}

func TestUpdateRowToSuccess(t *testing.T) {
	debtDueDate, _ := time.Parse("2006-01-02", "2023-12-31")
	bankSlip := newBankSlip(123, 1000.50, debtDueDate, "debt123", "John Doe", "john.doe@example.com", "file123", BankSlipStatusPending)

	bankSlip.Success()

	assert.Equal(t, bankSlip.Status, BankSlipStatusSuccess)
}
