package handler

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestMultipartFileHandlerSuit struct {
	suite.Suite
	multipartFileHandler FileHandler
}

func (testSuit *TestMultipartFileHandlerSuit) SetupTest() {
	testSuit.multipartFileHandler = NewMultipartFileHandler()
}

func TestMultipartFileHandler(t *testing.T) {
	suite.Run(t, new(TestMultipartFileHandlerSuit))
}

func (s *TestMultipartFileHandlerSuit) TestMultipartFileHandlerSuit_ShouldSaveCorrectly() {
	handler := NewMultipartFileHandler()

	mockFile := NewMockFile("test_file.csv", []byte("any_file"))

	savedFile, err := handler.SaveFile(mockFile)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), savedFile)

	// Verifica se o arquivo foi salvo corretamente
	savedFilePath := savedFile.Filepath()
	_, err = os.Stat(savedFilePath)
	assert.NoError(s.T(), err)

	// Limpa o arquivo salvo após o teste
	err = os.Remove(savedFilePath)
	os.RemoveAll("./uploads")
	assert.NoError(s.T(), err)
}

type MockFile struct {
	fileName string
	content  []byte
}

func NewMockFile(fileName string, content []byte) *MockFile {
	return &MockFile{
		fileName: fileName,
		content:  content,
	}
}

func (m *MockFile) FileName() string {
	return m.fileName
}

func (m *MockFile) Reader() io.Reader {
	return bytes.NewReader(m.content)
}

func (m *MockFile) Close() {
}
