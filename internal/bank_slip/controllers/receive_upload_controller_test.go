package bank_slip

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	bankSlipMocks "performatic-file-processor/internal/bank_slip/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type TestSuitReceiveUploadController struct {
	suite.Suite
	receiveUploadService *bankSlipMocks.ReceiveUploadServiceMock
	controller           *ReceiveUploadController
}

func (testSuit *TestSuitReceiveUploadController) SetupTest() {
	testSuit.receiveUploadService = new(bankSlipMocks.ReceiveUploadServiceMock)

	testSuit.controller = NewReceiveUploadController(
		testSuit.receiveUploadService,
	)
}

func TestReceiveUploadService(t *testing.T) {
	suite.Run(t, new(TestSuitReceiveUploadController))
}

func (s *TestSuitReceiveUploadController) TestReceiveUploadController_ShouldIgnoreWhenFormFromFileFails() {
	req := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewBuffer(nil))

	recorder := httptest.NewRecorder()

	s.controller.UploadBankSlipFileHandler(recorder, req)

	assert.Equal(s.T(), http.StatusInternalServerError, recorder.Code)

	expectedErrorResponse, _ := json.Marshal(map[string]string{"error": "Erro ao obter arquivo!"})
	assert.JSONEq(s.T(), string(expectedErrorResponse), recorder.Body.String())

	s.receiveUploadService.AssertNotCalled(s.T(), "Execute")
}

func (s *TestSuitReceiveUploadController) TestReceiveUploadController_ShouldCallServiceWhenFormFromFileSucceeds() {
	fileContent := []byte("any_file")
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile("file", "testfile.txt")
	part.Write(fileContent)

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType()) // O boundary é automaticamente incluído

	s.receiveUploadService.On("Execute", mock.Anything, mock.Anything).Return(nil)

	recorder := httptest.NewRecorder()
	s.controller.UploadBankSlipFileHandler(recorder, req)

	s.receiveUploadService.AssertCalled(s.T(), "Execute", mock.Anything, mock.Anything)
}
