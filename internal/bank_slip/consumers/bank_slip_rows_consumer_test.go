package bank_slip

import (
	"context"
	"sync"
	"testing"
	"time"

	bankSlipMocks "performatic-file-processor/internal/bank_slip/mocks"
	"performatic-file-processor/internal/messaging"
	"performatic-file-processor/internal/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TestSuitBankSlipRowsConsumer struct {
	suite.Suite
	mockMessageConsumer            *mocks.MessageConsumerMock
	mockProcessBankSlipRowsService *bankSlipMocks.ProcessBankSlipRowsServiceMock
	consumer                       *BankSlipRowsConsumer
}

func (testSuit *TestSuitBankSlipRowsConsumer) SetupTest() {
	testSuit.mockMessageConsumer = new(mocks.MessageConsumerMock)
	testSuit.mockProcessBankSlipRowsService = new(bankSlipMocks.ProcessBankSlipRowsServiceMock)

	testSuit.consumer = NewBankSlipRowsConsumer(
		testSuit.mockProcessBankSlipRowsService,
		testSuit.mockMessageConsumer,
		2,
	)
}

func TestReceiveUploadService(t *testing.T) {
	suite.Run(t, new(TestSuitBankSlipRowsConsumer))
}

func (s *TestSuitBankSlipRowsConsumer) TestBankSlipRowsConsumer_ShouldSendMessageToBeProcessed() {
	mockMessage := mocks.NewMessageMock()

	s.mockMessageConsumer.On("SubscribeInTopic", mock.Anything, "rows-to-process").Return(nil)
	s.mockMessageConsumer.On("Consume", mock.Anything, mock.Anything).Return(mockMessage, nil)

	s.mockProcessBankSlipRowsService.On("Execute", mock.Anything, mock.Anything).Return(nil).Twice()

	ctx, _ := context.WithTimeout(context.Background(), 500*time.Millisecond)
	chann := make(chan messaging.Message)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.consumer.Execute(ctx, chann)
	}()
	time.Sleep(200 * time.Millisecond)
	msg := <-chann
	ctx.Done()

	s.mockMessageConsumer.AssertCalled(s.T(), "Consume", ctx, "rows-to-process")
	s.mockProcessBankSlipRowsService.AssertNumberOfCalls(s.T(), "Execute", 2)

	s.Equal(mockMessage, msg)
}

func (s *TestSuitBankSlipRowsConsumer) TestBankSlipRowsConsumer_ShouldIgnoreWhenConsumerReturnsError() {
	s.mockMessageConsumer.On("SubscribeInTopic", mock.Anything, "rows-to-process").Return(nil)
	s.mockMessageConsumer.On("Consume", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	s.mockProcessBankSlipRowsService.On("Execute", mock.Anything, mock.Anything).Return(nil).Twice()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	chann := make(chan messaging.Message)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.consumer.Execute(ctx, chann)
	}()
	wg.Wait()

	s.mockMessageConsumer.AssertCalled(s.T(), "Consume", mock.Anything, "rows-to-process")
	s.mockProcessBankSlipRowsService.AssertNumberOfCalls(s.T(), "Execute", 2)
	select {
	case <-chann:
		s.T().Error("O canal deveria estar vazio, mas recebeu uma mensagem inesperada")
	default:
	}
}
