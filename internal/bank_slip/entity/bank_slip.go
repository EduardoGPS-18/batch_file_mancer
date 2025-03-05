package bank_slip

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

type BankSlipStatus string
type DebitId = string
type Existing = bool
type Success = bool

const (
	BankSlipStatusPending              BankSlipStatus = "PENDING"
	BankSlipStatusSuccess              BankSlipStatus = "SUCCESS"
	BankSlipStatusGenerateBillingError BankSlipStatus = "GENERATING_BILLING_ERROR"
	BankSlipStatusSendingEmailError    BankSlipStatus = "SENT_EMAIL_WITH_ERROR"
)

type BankSlipMap = map[DebitId]*BankSlip

type BankSlipRepository interface {
	UpdateMany(bankSlips ...*BankSlipMap) error
	InsertMany(bankSlips *BankSlipMap) (map[DebitId]Success, error)
}

type BankSlip struct {
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
		ErrorMessage:           nil,
	}
}

func NewBankSlipFromRow(fileId, data, header string) (*BankSlip, error) {
	rowItems := strings.Split(data, ",")
	headerItems := strings.Split(header, ",")

	userNamePosition := slices.IndexFunc(headerItems, func(s string) bool { return strings.Contains(s, "name") })
	governmentIdPosition := slices.IndexFunc(headerItems, func(s string) bool { return strings.Contains(s, "governmentId") })
	emailPosition := slices.IndexFunc(headerItems, func(s string) bool { return strings.Contains(s, "email") })
	amountPosition := slices.IndexFunc(headerItems, func(s string) bool { return strings.Contains(s, "debtAmount") })
	dueDatePosition := slices.IndexFunc(headerItems, func(s string) bool { return strings.Contains(s, "debtDueDate") })
	debtIdPosition := slices.IndexFunc(headerItems, func(s string) bool { return strings.Contains(s, "debtId") })

	if len(fileId) == 0 {
		return nil, fmt.Errorf("file metadata must be not empty")
	}
	someVariableIsMissing := governmentIdPosition == -1 || userNamePosition == -1 || emailPosition == -1 || amountPosition == -1 || dueDatePosition == -1 || debtIdPosition == -1
	if someVariableIsMissing {
		return nil, fmt.Errorf("error is missing some field (file id: %s)", fileId)
	}
	if len(rowItems) != len(headerItems) {
		return nil, fmt.Errorf("error rowItems and headerItems length are different (file id: %s)", fileId)
	}
	governmentId, err := strconv.Atoi(rowItems[governmentIdPosition])
	if err != nil {
		return nil, fmt.Errorf("error converting governmentId to int %s Position: %s (file id: %s)", string(rowItems[governmentIdPosition]), fmt.Sprint(governmentIdPosition), fileId)
	}
	debtAmount, err := strconv.ParseFloat(rowItems[amountPosition], 64)
	if err != nil {
		return nil, fmt.Errorf("error converting debtAmount to float64 %s Position: %s (file id: %s)", string(rowItems[amountPosition]), fmt.Sprint(amountPosition), fileId)
	}
	debtDueDate, err := time.Parse("2006-01-02", rowItems[dueDatePosition])
	if err != nil {
		return nil, fmt.Errorf("error converting debtDueDate to time.Time %s Position: %s (file id: %s)", string(rowItems[dueDatePosition]), fmt.Sprint(dueDatePosition), fileId)
	}

	return newBankSlip(
		governmentId,
		debtAmount,
		debtDueDate,
		rowItems[debtIdPosition],
		rowItems[userNamePosition],
		rowItems[emailPosition],
		fileId,
		BankSlipStatusPending,
	), nil
}

func (bankSlip *BankSlip) ErrorGeneratingBilling(errorMessage string) {
	bankSlip.ErrorMessage = &errorMessage
	bankSlip.Status = BankSlipStatusGenerateBillingError
}

func (bankSlip *BankSlip) ErrorSendingEmail(errorMessage string) {
	bankSlip.ErrorMessage = &errorMessage
	bankSlip.Status = BankSlipStatusSendingEmailError
}

func (bankSlip *BankSlip) Success() {
	bankSlip.Status = BankSlipStatusSuccess
}
