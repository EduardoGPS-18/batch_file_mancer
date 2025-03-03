package bank_slip

import (
	"bytes"
	"io"
	"reflect"
	"testing"

	bankSlipEntities "performatic-file-processor/internal/bank_slip/entity"
	bankSlipMocks "performatic-file-processor/internal/bank_slip/mocks"
	"performatic-file-processor/internal/handler"
	sharedMocks "performatic-file-processor/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TestSuitReceiveUploadService struct {
	suite.Suite
	mockBankSlipRepo         *bankSlipMocks.BankSlipRepositoryMock
	mockBankSlipFileRepo     *bankSlipMocks.BankSlipFileMetadataRepositoryMock
	mockMultipartFileHandler *sharedMocks.FileHandlerMock
	mockMessageProducer      *sharedMocks.MessageProducerMock
	service                  *ReceiveUploadService
}

func (testSuit *TestSuitReceiveUploadService) SetupTest() {
	testSuit.mockBankSlipRepo = new(bankSlipMocks.BankSlipRepositoryMock)
	testSuit.mockBankSlipFileRepo = new(bankSlipMocks.BankSlipFileMetadataRepositoryMock)
	testSuit.mockMultipartFileHandler = new(sharedMocks.FileHandlerMock)
	testSuit.mockMessageProducer = new(sharedMocks.MessageProducerMock)

	testSuit.service = NewReceiveUploadService(
		testSuit.mockBankSlipRepo,
		testSuit.mockBankSlipFileRepo,
		testSuit.mockMultipartFileHandler,
		testSuit.mockMessageProducer,
		len("headerData"),
		2,
	)
}

func TestReceiveUploadService(t *testing.T) {
	suite.Run(t, new(TestSuitReceiveUploadService))
}

func (suit *TestSuitReceiveUploadService) TestReceiveUploadService_ShouldReturnErrorIfInsertMetadataFails() {
	fileContent := bytes.NewBufferString("headerData\nrow1\nrow2\n").Bytes()
	fileName := "testfile.txt"
	file, fileHeaders, err := sharedMocks.CreateMultipartFileMock(fileName, fileContent)
	if err != nil {
		panic(err)
	}

	suit.mockBankSlipFileRepo.On("Insert", mock.Anything).Return(assert.AnError).Once()

	err = suit.service.Execute(file, fileHeaders)
	assert.Error(suit.T(), err)

	suit.mockBankSlipFileRepo.AssertCalled(suit.T(), "Insert", mock.MatchedBy(func(bankSlipFile *bankSlipEntities.BankSlipFileMetadata) bool {
		return assert.Equal(suit.T(), fileName, bankSlipFile.FileName)
	}))
}

func (suit *TestSuitReceiveUploadService) TestReceiveUploadService_ShouldReturnErrorIfSaveFileFails() {
	fileContent := bytes.NewBufferString("headerData\nrow1\nrow2\n").Bytes()
	fileName := "testfile.txt"
	file, fileHeaders, err := sharedMocks.CreateMultipartFileMock(fileName, fileContent)
	if err != nil {
		panic(err)
	}

	suit.mockBankSlipFileRepo.On("Insert", mock.Anything).Return(nil).Once()
	suit.mockMultipartFileHandler.On("SaveFile", mock.Anything).Return(nil, assert.AnError).Once()

	err = suit.service.Execute(file, fileHeaders)
	assert.Error(suit.T(), err)

	suit.mockBankSlipFileRepo.AssertCalled(suit.T(), "Insert", mock.MatchedBy(func(bankSlipFile *bankSlipEntities.BankSlipFileMetadata) bool {
		return assert.Equal(suit.T(), fileName, bankSlipFile.FileName)
	}))
	suit.mockMultipartFileHandler.AssertCalled(suit.T(), "SaveFile", handler.NewMultipartFile(file, fileHeaders))
}

func (suit *TestSuitReceiveUploadService) TestReceiveUploadService_ShouldIgnoreWhenReadHeaderThrowsNotEOF() {
	fileContent := bytes.NewBufferString("headerData\nrow1,row1\nrow2,row2").Bytes()
	fileName := "testfile.txt"
	file, fileHeaders, err := sharedMocks.CreateMultipartFileMock(fileName, fileContent)
	if err != nil {
		panic(err)
	}

	mockSavedFile := sharedMocks.NewSavedFileMock()
	mockedReader := sharedMocks.NewReaderMock()

	suit.mockBankSlipFileRepo.On("Insert", mock.Anything).Run(func(arg mock.Arguments) {
		arg.Get(0).(*bankSlipEntities.BankSlipFileMetadata).ID = "any_id"
	}).Return(nil).Once()
	suit.mockMultipartFileHandler.On("SaveFile", mock.Anything).Return(mockSavedFile, nil).Once()
	suit.mockMessageProducer.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockedReader.On("Read", mock.Anything).Return(0, assert.AnError).Once()
	mockSavedFile.On("Open").Return(mockedReader).Once()
	mockSavedFile.On("Delete").Return(nil).Once()

	err = suit.service.Execute(file, fileHeaders)
	assert.Error(suit.T(), err)

	suit.mockBankSlipFileRepo.AssertCalled(suit.T(), "Insert", mock.MatchedBy(func(bankSlipFile *bankSlipEntities.BankSlipFileMetadata) bool {
		return assert.Equal(suit.T(), fileName, bankSlipFile.FileName)
	}))
	suit.mockMultipartFileHandler.AssertCalled(suit.T(), "SaveFile", handler.NewMultipartFile(file, fileHeaders))
	suit.mockMessageProducer.AssertNotCalled(suit.T(), "Publish")
}

func (suit *TestSuitReceiveUploadService) TestReceiveUploadService_ShouldIgnoreWhenFailsReadingHeader() {
	fileContent := bytes.NewBufferString("headerData\nrow1,row1\nrow2,row2").Bytes()
	fileName := "testfile.txt"
	file, fileHeaders, err := sharedMocks.CreateMultipartFileMock(fileName, fileContent)
	if err != nil {
		panic(err)
	}

	mockSavedFile := sharedMocks.NewSavedFileMock()
	mockedReader := sharedMocks.NewReaderMock()

	suit.mockBankSlipFileRepo.On("Insert", mock.Anything).Run(func(arg mock.Arguments) {
		arg.Get(0).(*bankSlipEntities.BankSlipFileMetadata).ID = "any_id"
	}).Return(nil).Once()
	suit.mockMultipartFileHandler.On("SaveFile", mock.Anything).Return(mockSavedFile, nil).Once()
	suit.mockMessageProducer.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockedReader.On("Read", mock.Anything).Return(5, assert.AnError).Once()
	mockSavedFile.On("Open").Return(mockedReader).Once()
	mockSavedFile.On("Delete").Return(nil).Once()

	err = suit.service.Execute(file, fileHeaders)
	assert.Error(suit.T(), err)

	suit.mockBankSlipFileRepo.AssertCalled(suit.T(), "Insert", mock.MatchedBy(func(bankSlipFile *bankSlipEntities.BankSlipFileMetadata) bool {
		return assert.Equal(suit.T(), fileName, bankSlipFile.FileName)
	}))
	suit.mockMultipartFileHandler.AssertCalled(suit.T(), "SaveFile", handler.NewMultipartFile(file, fileHeaders))
	suit.mockMessageProducer.AssertNotCalled(suit.T(), "Publish")
}
