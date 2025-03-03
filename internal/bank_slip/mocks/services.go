package bank_slip

import (
	"context"
	"mime/multipart"
	"performatic-file-processor/internal/messaging"

	"github.com/stretchr/testify/mock"
)

type ProcessBankSlipRowsServiceMock struct {
	mock.Mock
}

func (s *ProcessBankSlipRowsServiceMock) Execute(
	context context.Context,
	messagesChannel chan messaging.Message,
) {
	s.Called()
}

type ReceiveUploadServiceMock struct {
	mock.Mock
}

func (s *ReceiveUploadServiceMock) Execute(
	file multipart.File,
	fileHeader *multipart.FileHeader,
) error {
	args := s.Called()
	return args.Error(1)
}
