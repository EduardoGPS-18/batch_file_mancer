package bank_slip

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBankSlipFileMetadata(t *testing.T) {
	fileName := "test_file.txt"
	bankSlipFile := NewBankSlipFileMetadata(fileName)

	assert.Equal(t, fileName, bankSlipFile.FileName)
}
