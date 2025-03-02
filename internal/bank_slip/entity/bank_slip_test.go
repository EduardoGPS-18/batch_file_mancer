package bank_slip

import (
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

func TestNewBankSlipFromRow_Error(t *testing.T) {
	header := "name,governmentId,email,debtAmount,debtDueDate,debtId"
	data := "John Doe,abc,john.doe@example.com,1000.50,2023-12-31,debt123"
	fileMetadataId := "file123"

	bankSlip, err := NewBankSlipFromRow(fileMetadataId, data, header)
	assert.Error(t, err)
	assert.Nil(t, bankSlip)
}

func TestUpdateRowToError(t *testing.T) {
	debtDueDate, _ := time.Parse("2006-01-02", "2023-12-31")
	bankSlip := newBankSlip(123, 1000.50, debtDueDate, "debt123", "John Doe", "john.doe@example.com", "file123", BankSlipStatusPending)

	errorMessage := "Some error occurred"
	updatedBankSlip, err := bankSlip.UpdateRowToError(errorMessage)
	assert.NoError(t, err)

	assert.Equal(t, BankSlipStatusError, updatedBankSlip.Status)
	assert.Equal(t, &errorMessage, updatedBankSlip.ErrorMessage)
}
