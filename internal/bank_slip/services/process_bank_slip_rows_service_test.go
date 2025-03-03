package bank_slip

import (
	"context"
	bankSlipEntities "performatic-file-processor/internal/bank_slip/entity"
	bankSlipMocks "performatic-file-processor/internal/bank_slip/mocks"
	"performatic-file-processor/internal/messaging"
	sharedMocks "performatic-file-processor/internal/mocks"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TestSuit struct {
	suite.Suite
	mockBankSlipFileRepository *bankSlipMocks.BankSlipFileMetadataRepositoryMock
	mockBankSlipRepository     *bankSlipMocks.BankSlipRepositoryMock
	service                    *ProcessBankSlipRowsService
}

func (s *TestSuit) SetupTest() {
	s.mockBankSlipFileRepository = new(bankSlipMocks.BankSlipFileMetadataRepositoryMock)
	s.mockBankSlipRepository = new(bankSlipMocks.BankSlipRepositoryMock)
	s.service = NewProcessBankSlipRowsService(
		s.mockBankSlipFileRepository,
		s.mockBankSlipRepository,
	)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(TestSuit))
}

func (s *TestSuit) TestProcessBankSlipRowsService_ShouldDoNothingWhenFailToConvertMessageData() {
	messagesChannel := make(chan messaging.Message, 1)
	message := sharedMocks.NewKafkaMessageMock()
	messagesChannel <- message

	message.On("Data").Return(nil, assert.AnError).Once()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.service.Execute(context.Background(), messagesChannel)
	}()
	close(messagesChannel)
	wg.Wait()

	message.AssertCalled(s.T(), "Data")
	message.AssertNotCalled(s.T(), "Commit")
	s.mockBankSlipRepository.AssertNotCalled(s.T(), "GetExistingByDebitIds")
	s.mockBankSlipRepository.AssertNotCalled(s.T(), "InsertMany")
}

func (s *TestSuit) TestProcessBankSlipRowsService_ShouldDoNothingWhenFailCreatingBankSlipEntity() {
	message := sharedMocks.NewKafkaMessageMock()

	messageWithHeaderAndDataWithDiferentLength := map[string]any{
		"header": "name,governmentId,email,debtAmount,debtDueDate,debtId",
		"data":   "John Doe,john.doe@example.com,1000.50,2023-12-31,debt123",
		"fileId": "fileId",
	}
	message.On("Data").Return(messageWithHeaderAndDataWithDiferentLength, nil).Once()
	message.On("Commit")

	mockedData := map[bankSlipEntities.DebitId]bankSlipEntities.Existing{}
	s.mockBankSlipRepository.On("GetExistingByDebitIds", mock.Anything).Return(mockedData, nil).Once()
	s.mockBankSlipRepository.On("InsertMany", mock.Anything).Return(nil).Once()

	messagesChannel := make(chan messaging.Message, 1)
	messagesChannel <- message

	message.On("Data").Return(nil, assert.AnError).Once()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.service.Execute(context.Background(), messagesChannel)
	}()
	close(messagesChannel)
	wg.Wait()

	message.AssertCalled(s.T(), "Data")
	message.AssertNotCalled(s.T(), "Commit")
	s.mockBankSlipRepository.AssertNotCalled(s.T(), "GetExistingByDebitIds")
	s.mockBankSlipRepository.AssertNotCalled(s.T(), "InsertMany")
}

func (s *TestSuit) TestProcessBankSlipRowsService_ShouldDoNothingWhenFailGettingExistingDebits() {
	message := sharedMocks.NewKafkaMessageMock()

	messageWithHeaderAndDataWithDiferentLength := map[string]any{
		"header": "name,governmentId,email,debtAmount,debtDueDate,debtId",
		"data":   "John Doe,john.doe@example.com,1000.50,2023-12-31,debt123",
		"fileId": "fileId",
	}
	message.On("Data").Return(messageWithHeaderAndDataWithDiferentLength, nil).Once()
	message.On("Commit")

	s.mockBankSlipRepository.On("GetExistingByDebitIds", mock.Anything).Return(nil, assert.AnError).Once()
	s.mockBankSlipRepository.On("InsertMany", mock.Anything).Return(nil).Once()

	messagesChannel := make(chan messaging.Message, 1)
	messagesChannel <- message

	message.On("Data").Return(nil, assert.AnError).Once()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.service.Execute(context.Background(), messagesChannel)
	}()
	close(messagesChannel)
	wg.Wait()

	message.AssertCalled(s.T(), "Data")
	message.AssertNotCalled(s.T(), "Commit")
	s.mockBankSlipRepository.AssertNotCalled(s.T(), "GetExistingByDebitIds")
	s.mockBankSlipRepository.AssertNotCalled(s.T(), "InsertMany")
}

func (s *TestSuit) TestProcessBankSlipRowsService_ShouldDoNothingIfTheresNtDebitToInsert() {
	message := sharedMocks.NewKafkaMessageMock()

	messageWithHeaderAndDataWithDiferentLength := map[string]any{
		"header": "name,governmentId,email,debtAmount,debtDueDate,debtId",
		"data":   "",
		"fileId": "fileId",
	}
	message.On("Data").Return(messageWithHeaderAndDataWithDiferentLength, nil).Once()
	message.On("Commit")

	mockedData := map[bankSlipEntities.DebitId]bankSlipEntities.Existing{
		"debt123": true,
	}
	s.mockBankSlipRepository.On("GetExistingByDebitIds", mock.Anything).Return(mockedData, nil).Once()
	s.mockBankSlipRepository.On("InsertMany", mock.Anything).Return(nil).Once()

	messagesChannel := make(chan messaging.Message, 1)
	messagesChannel <- message

	message.On("Data").Return(nil, assert.AnError).Once()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.service.Execute(context.Background(), messagesChannel)
	}()
	close(messagesChannel)
	wg.Wait()

	message.AssertNotCalled(s.T(), "Commit")
	s.mockBankSlipRepository.AssertNotCalled(s.T(), "GetExistingByDebitIds")
	s.mockBankSlipRepository.AssertNotCalled(s.T(), "InsertMany")
}

func (s *TestSuit) TestProcessBankSlipRowsService_ShouldSaveWithErrorIfDebitAlreadyExists() {
	message := sharedMocks.NewKafkaMessageMock()

	messageWithHeaderAndDataWithDiferentLength := map[string]any{
		"header": "name,governmentId,email,debtAmount,debtDueDate,debtId",
		"data":   "John Doe,555,john.doe@example.com,1000.50,2023-12-31,debt123",
		"fileId": "fileId",
	}
	message.On("Data").Return(messageWithHeaderAndDataWithDiferentLength, nil).Once()
	message.On("Commit")

	mockedData := map[bankSlipEntities.DebitId]bankSlipEntities.Existing{
		"debt123": true,
	}
	s.mockBankSlipRepository.On("GetExistingByDebitIds", mock.Anything).Return(mockedData, nil).Once()

	errorTxt := "Debt already exists"
	expected := &bankSlipEntities.BankSlip{
		ID:                     0,
		UserName:               "John Doe",
		GovernmentId:           555,
		UserEmail:              "john.doe@example.com",
		BankSlipFileMetadataId: "fileId",
		ErrorMessage:           &errorTxt,
		Status:                 bankSlipEntities.BankSlipStatusError,
		DebtAmount:             1000.50,
		DebtDueDate:            time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
		DebtId:                 "debt123",
	}

	s.mockBankSlipRepository.On("InsertMany", mock.Anything).Return(nil).Once()

	messagesChannel := make(chan messaging.Message, 1)
	messagesChannel <- message

	message.On("Data").Return(nil, assert.AnError).Once()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.service.Execute(context.Background(), messagesChannel)
	}()
	close(messagesChannel)
	wg.Wait()

	message.AssertCalled(s.T(), "Data")
	message.AssertCalled(s.T(), "Commit")
	s.mockBankSlipRepository.AssertCalled(s.T(), "GetExistingByDebitIds", mock.MatchedBy(func(ids []string) bool {
		expected := []string{"'debt123'"}
		return assert.ElementsMatch(s.T(), ids, expected)
	}))
	s.mockBankSlipRepository.AssertCalled(s.T(), "InsertMany", mock.MatchedBy(func(m map[bankSlipEntities.DebitId]*bankSlipEntities.BankSlip) bool {
		actual, exists := m["debt123"]
		return exists && assert.Equal(s.T(), expected, actual)
	}))
	message.AssertCalled(s.T(), "Commit")
}

func (s *TestSuit) TestProcessBankSlipRowsService_ShouldNotCommitMessageWhenInsertFails() {
	message := sharedMocks.NewKafkaMessageMock()

	message.On("Data").Return(map[string]any{
		"header": "name,governmentId,email,debtAmount,debtDueDate,debtId",
		"data":   "John Doe,123,john.doe@example.com,1000.50,2023-12-31,debt123",
		"fileId": "fileId",
	}, nil).Once()
	message.On("Commit")

	mockedData := map[bankSlipEntities.DebitId]bankSlipEntities.Existing{}
	s.mockBankSlipRepository.On("GetExistingByDebitIds", mock.Anything).Return(mockedData, nil).Once()
	s.mockBankSlipRepository.On("InsertMany", mock.Anything).Return(assert.AnError).Once()

	messagesChannel := make(chan messaging.Message, 1)
	messagesChannel <- message
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.service.Execute(context.Background(), messagesChannel)
	}()
	close(messagesChannel)
	wg.Wait()

	s.mockBankSlipRepository.AssertCalled(s.T(), "GetExistingByDebitIds", mock.MatchedBy(func(ids []string) bool {
		expected := []string{"'debt123'"}
		return assert.ElementsMatch(s.T(), expected, ids)
	}))

	expected := &bankSlipEntities.BankSlip{
		ID:                     0,
		UserName:               "John Doe",
		GovernmentId:           123,
		UserEmail:              "john.doe@example.com",
		BankSlipFileMetadataId: "fileId",
		ErrorMessage:           nil,
		Status:                 bankSlipEntities.BankSlipStatusPending,
		DebtAmount:             1000.50,
		DebtDueDate:            time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
		DebtId:                 "debt123",
	}
	s.mockBankSlipRepository.AssertCalled(s.T(), "InsertMany", mock.MatchedBy(func(m map[bankSlipEntities.DebitId]*bankSlipEntities.BankSlip) bool {
		actual, exists := m["debt123"]
		return exists && assert.Equal(s.T(), expected, actual)
	}))
	message.AssertNotCalled(s.T(), "Commit")
}

func (s *TestSuit) TestProcessBankSlipRowsService_ShouldProcessSuccessfullyBankSlipRows() {
	message := sharedMocks.NewKafkaMessageMock()

	message.On("Data").Return(map[string]any{
		"header": "name,governmentId,email,debtAmount,debtDueDate,debtId",
		"data":   "John Doe,123,john.doe@example.com,1000.50,2023-12-31,debt123",
		"fileId": "fileId",
	}, nil).Once()
	message.On("Commit")

	mockedData := map[bankSlipEntities.DebitId]bankSlipEntities.Existing{}
	s.mockBankSlipRepository.On("GetExistingByDebitIds", mock.Anything).Return(mockedData, nil).Once()
	s.mockBankSlipRepository.On("InsertMany", mock.Anything).Return(nil).Once()

	messagesChannel := make(chan messaging.Message, 1)
	messagesChannel <- message
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.service.Execute(context.Background(), messagesChannel)
	}()
	close(messagesChannel)
	wg.Wait()

	s.mockBankSlipRepository.AssertCalled(s.T(), "GetExistingByDebitIds", mock.MatchedBy(func(ids []string) bool {
		expected := []string{"'debt123'"}
		return assert.ElementsMatch(s.T(), expected, ids)
	}))

	expected := &bankSlipEntities.BankSlip{
		ID:                     0,
		UserName:               "John Doe",
		GovernmentId:           123,
		UserEmail:              "john.doe@example.com",
		BankSlipFileMetadataId: "fileId",
		ErrorMessage:           nil,
		Status:                 bankSlipEntities.BankSlipStatusPending,
		DebtAmount:             1000.50,
		DebtDueDate:            time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
		DebtId:                 "debt123",
	}

	s.mockBankSlipRepository.AssertCalled(s.T(), "InsertMany", mock.MatchedBy(func(m map[bankSlipEntities.DebitId]*bankSlipEntities.BankSlip) bool {
		actual, exists := m["debt123"]
		return exists && assert.Equal(s.T(), expected, actual)
	}))
	message.AssertCalled(s.T(), "Commit")
}

func (s *TestSuit) TestProcessBankSlipRowsService_ShouldProcessSuccessfullyWhenFirstReceivedLineIsBanlkBankSlipRows() {
	message := sharedMocks.NewKafkaMessageMock()

	message.On("Data").Return(map[string]any{
		"header": "name,governmentId,email,debtAmount,debtDueDate,debtId",
		"data":   "\nJohn Doe,123,john.doe@example.com,1000.50,2023-12-31,debt123",
		"fileId": "fileId",
	}, nil).Once()
	message.On("Commit")

	mockedData := map[bankSlipEntities.DebitId]bankSlipEntities.Existing{}
	s.mockBankSlipRepository.On("GetExistingByDebitIds", mock.Anything).Return(mockedData, nil).Once()
	s.mockBankSlipRepository.On("InsertMany", mock.Anything).Return(nil).Once()

	messagesChannel := make(chan messaging.Message, 1)
	messagesChannel <- message
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.service.Execute(context.Background(), messagesChannel)
	}()
	close(messagesChannel)
	wg.Wait()

	s.mockBankSlipRepository.AssertCalled(s.T(), "GetExistingByDebitIds", mock.MatchedBy(func(ids []string) bool {
		expected := []string{"'debt123'"}
		return assert.ElementsMatch(s.T(), expected, ids)
	}))

	expected := &bankSlipEntities.BankSlip{
		ID:                     0,
		UserName:               "John Doe",
		GovernmentId:           123,
		UserEmail:              "john.doe@example.com",
		BankSlipFileMetadataId: "fileId",
		ErrorMessage:           nil,
		Status:                 bankSlipEntities.BankSlipStatusPending,
		DebtAmount:             1000.50,
		DebtDueDate:            time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
		DebtId:                 "debt123",
	}

	s.mockBankSlipRepository.AssertCalled(s.T(), "InsertMany", mock.MatchedBy(func(m map[bankSlipEntities.DebitId]*bankSlipEntities.BankSlip) bool {
		actual, exists := m["debt123"]
		return exists && assert.Equal(s.T(), expected, actual)
	}))
	message.AssertCalled(s.T(), "Commit")
}

func (s *TestSuit) TestProcessBankSlipRowsService_ShouldProcessOnlyValidMessagesWhenManyRowIsProvided() {
	message := sharedMocks.NewKafkaMessageMock()

	message.On("Data").Return(map[string]any{
		"header": "name,governmentId,email,debtAmount,debtDueDate,debtId",
		"data":   "Mary Doe,987,mary.doe@example.com,5021.50,2023-12-31,debt543\nJohn Doe,abc,john.doe@example.com,1000.50,2023-12-31,debt123\n",
		"fileId": "fileId",
	}, nil).Once()
	message.On("Commit")

	mockedData := map[bankSlipEntities.DebitId]bankSlipEntities.Existing{}
	s.mockBankSlipRepository.On("GetExistingByDebitIds", mock.Anything).Return(mockedData, nil).Once()
	s.mockBankSlipRepository.On("InsertMany", mock.Anything).Return(nil).Once()

	messagesChannel := make(chan messaging.Message, 1)
	messagesChannel <- message
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.service.Execute(context.Background(), messagesChannel)
	}()
	close(messagesChannel)
	wg.Wait()

	s.mockBankSlipRepository.AssertCalled(s.T(), "GetExistingByDebitIds", mock.MatchedBy(func(ids []string) bool {
		expected := []string{"'debt543'"}
		return assert.ElementsMatch(s.T(), expected, ids)
	}))

	expected := &bankSlipEntities.BankSlip{
		ID:                     0,
		UserName:               "Mary Doe",
		GovernmentId:           987,
		UserEmail:              "mary.doe@example.com",
		BankSlipFileMetadataId: "fileId",
		ErrorMessage:           nil,
		Status:                 bankSlipEntities.BankSlipStatusPending,
		DebtAmount:             5021.50,
		DebtDueDate:            time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
		DebtId:                 "debt543",
	}
	s.mockBankSlipRepository.AssertCalled(s.T(), "InsertMany", mock.MatchedBy(func(m map[bankSlipEntities.DebitId]*bankSlipEntities.BankSlip) bool {
		actual, exists := m["debt543"]
		return exists && assert.Equal(s.T(), expected, actual)
	}))
	message.AssertCalled(s.T(), "Commit")
}

func (s *TestSuit) TestProcessBankSlipRowsService_ShouldExitWhenContextIsDone() {
	mockedData := map[bankSlipEntities.DebitId]bankSlipEntities.Existing{}
	s.mockBankSlipRepository.On("GetExistingByDebitIds", mock.Anything).Return(mockedData, nil).Once()
	s.mockBankSlipRepository.On("InsertMany", mock.Anything).Return(nil).Once()

	messagesChannel := make(chan messaging.Message, 1)
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.service.Execute(ctx, messagesChannel)
	}()
	cancel()
	wg.Wait()
}
