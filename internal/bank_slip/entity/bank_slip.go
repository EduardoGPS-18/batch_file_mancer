package bank_slip

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

type BankSlipStatus string
type DebitId = string
type Existing = bool

const (
	BankSlipStatusPending       BankSlipStatus = "PENDING"
	BankSlipStatusError         BankSlipStatus = "ERROR"
	BankSlipStatusSlipGenerated BankSlipStatus = "GENERATED"
	BankSlipStatusEmailSent     BankSlipStatus = "SENT"
)

type BankSlipRepository interface {
	InsertMany(bankSlips map[DebitId]*BankSlip) error
	GetExistingByDebitIds(debitIds []string) (map[DebitId]Existing, error)
}

type BankSlip struct {
	ID                     int
	DebtId                 string
	DebtAmount             float64
	DebtDueDate            time.Time
	GovernmentId           int
	UserName               string
	UserEmail              string
	BankSlipFileMetadataId string
	ErrorMessage           *string
	Status                 BankSlipStatus
}

func newBankSlip(governmentId int, debtAmount float64, debtDueDate time.Time, debtId, userName, userEmail, bankSlipFileMetadataId string, status BankSlipStatus) *BankSlip {
	return &BankSlip{
		DebtId:                 debtId,
		DebtAmount:             debtAmount,
		DebtDueDate:            debtDueDate,
		GovernmentId:           governmentId,
		UserName:               userName,
		UserEmail:              userEmail,
		BankSlipFileMetadataId: bankSlipFileMetadataId,
		Status:                 status,
	}
}

func NewBankSlipFromRow(fileMetadataId, data, header string) (*BankSlip, error) {
	rowItems := strings.Split(data, ",")
	headerItems := strings.Split(header, ",")

	userNamePosition := slices.IndexFunc(headerItems, func(s string) bool { return strings.Contains(s, "name") })
	governmentIdPosition := slices.IndexFunc(headerItems, func(s string) bool { return strings.Contains(s, "governmentId") })
	emailPosition := slices.IndexFunc(headerItems, func(s string) bool { return strings.Contains(s, "email") })
	amountPosition := slices.IndexFunc(headerItems, func(s string) bool { return strings.Contains(s, "debtAmount") })
	dueDatePosition := slices.IndexFunc(headerItems, func(s string) bool { return strings.Contains(s, "debtDueDate") })
	debtIdPosition := slices.IndexFunc(headerItems, func(s string) bool { return strings.Contains(s, "debtId") })

	if len(fileMetadataId) == 0 {
		return nil, errors.New("file metadata must be not empty")
	}
	someVariableIsMissing := governmentIdPosition == -1 || userNamePosition == -1 || emailPosition == -1 || amountPosition == -1 || dueDatePosition == -1 || debtIdPosition == -1
	if someVariableIsMissing {
		return nil, errors.New("error is missing some field")
	}
	if len(rowItems) != len(headerItems) {
		return nil, errors.New("error rowItems and headerItems length are different")
	}
	governmentId, err := strconv.Atoi(rowItems[governmentIdPosition])
	if err != nil {
		return nil, errors.New("error converting governmentId to int " + string(rowItems[governmentIdPosition]) + " Position: " + fmt.Sprint(governmentIdPosition))
	}
	debtAmount, err := strconv.ParseFloat(rowItems[amountPosition], 64)
	if err != nil {
		return nil, errors.New("error converting debtAmount to float64 " + string(rowItems[amountPosition]) + " Position: " + fmt.Sprint(amountPosition))
	}
	debtDueDate, err := time.Parse("2006-01-02", rowItems[dueDatePosition])
	if err != nil {
		return nil, errors.New("error converting debtDueDate to time.Time " + string(rowItems[dueDatePosition]) + " Position: " + fmt.Sprint(dueDatePosition))
	}

	return newBankSlip(
		governmentId,
		debtAmount,
		debtDueDate,
		rowItems[debtIdPosition],
		rowItems[userNamePosition],
		rowItems[emailPosition],
		fileMetadataId,
		BankSlipStatusPending,
	), nil
}

func (bankSlip *BankSlip) UpdateRowToError(errorMessage string) {
	bankSlip.ErrorMessage = &errorMessage
	bankSlip.Status = BankSlipStatusError
}
