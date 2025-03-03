package bank_slip

import (
	"context"
	"testing"

	bankSlipMocks "performatic-file-processor/internal/bank_slip/mocks"
	"performatic-file-processor/internal/messaging"
	"performatic-file-processor/internal/mocks"
	sharedMocks "performatic-file-processor/internal/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TestSuitBankSlipRowsConsumer struct {
	suite.Suite
	mockBankSlipRepo         *bankSlipMocks.BankSlipRepositoryMock
	mockBankSlipFileRepo     *bankSlipMocks.BankSlipFileMetadataRepositoryMock
	mockMultipartFileHandler *sharedMocks.FileHandlerMock
	mockMessageProducer      *sharedMocks.MessageProducerMock
	service                  *ReceiveUploadService
}

func (testSuit *TestSuitBankSlipRowsConsumer) SetupTest() {
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
	suite.Run(t, new(TestSuitBankSlipRowsConsumer))
}
func (testSuit *TestSuitBankSlipRowsConsumer) TestExecuteSuccess() {
	mockMessageConsumer := new(mocks.MessageConsumerMock)
	mockProcessBankSlipRowsService := new(bankSlipMocks.ProcessBankSlipRowsServiceMock)

	consumer := NewBankSlipRowsConsumer(
		mockProcessBankSlipRowsService,
		mockMessageConsumer,
		2,
	)

	mockMessage := mocks.NewMessageMock()

	mockMessageConsumer.On("SubscribeInTopic", mock.Anything, "rows-to-process").Return(nil)
	mockMessageConsumer.On("Consume", mock.Anything, "rows-to-process").Return(mockMessage, nil).Once()

	mockProcessBankSlipRowsService.On("Execute", mock.Anything, mock.Anything).Return(nil).Once()

	go consumer.Execute()

	mockMessageConsumer.AssertExpectations(testSuit.T())
	mockProcessBankSlipRowsService.AssertExpectations(testSuit.T())
}

func (testSuit *TestSuitBankSlipRowsConsumer) TestExecuteConsumeError() {
	mockMessageConsumer := new(sharedMocks.MessageConsumerMock)
	mockProcessBankSlipRowsService := new(bankSlipMocks.ProcessBankSlipRowsServiceMock)

	consumer := NewBankSlipRowsConsumer(
		mockProcessBankSlipRowsService,
		mockMessageConsumer,
		2,
	)

	mockMessageConsumer.On("SubscribeInTopic", mock.Anything, "rows-to-process").Return(nil)
	mockMessageConsumer.On("Consume", mock.Anything, "rows-to-process").Return(messaging.Message{}, context.Canceled).Once()

	go consumer.Execute()

	mockMessageConsumer.AssertExpectations(testSuit.T())
	mockProcessBankSlipRowsService.AssertExpectations(testSuit.T())
}
