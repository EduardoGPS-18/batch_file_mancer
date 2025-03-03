package mocks

import (
	"bytes"
	"io"
	"mime/multipart"
	fileHandler "performatic-file-processor/internal/handler"

	"github.com/stretchr/testify/mock"
)

type FileMock struct {
	mock.Mock
}

func NewFileMock() *FileMock {
	return &FileMock{}
}

func (m *FileMock) FileName() string {
	args := m.Called()
	return args.String(0)
}

func (m *FileMock) Close() {
	m.Called()
}

func (m *FileMock) Reader() io.Reader {
	args := m.Called()
	return args.Get(0).(io.Reader)
}

type FileHandlerMock struct {
	mock.Mock
}

func NewFileHandlerMock() *FileHandlerMock {
	return &FileHandlerMock{}
}

func (m *FileHandlerMock) SaveFile(file fileHandler.File) (savedFile fileHandler.SavedFile, err error) {
	args := m.Called(file)
	var arg0 fileHandler.SavedFile = nil
	if args.Get(0) != nil {
		arg0 = args.Get(0).(fileHandler.SavedFile)
	}

	return arg0, args.Error(1)
}

func CreateMultipartFileMock(fileName string, fileContent []byte) (multipart.File, *multipart.FileHeader, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Criamos a parte do arquivo
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, nil, err
	}

	// Escrevemos o conteúdo no campo do formulário
	_, err = part.Write(fileContent)
	if err != nil {
		return nil, nil, err
	}

	// Fechamos o writer para finalizar a escrita no buffer
	writer.Close()

	// Criamos um novo leitor multipart para processar os dados gerados
	reader := multipart.NewReader(body, writer.Boundary())

	// Extraímos o formulário e definimos um limite de 10MB
	form, err := reader.ReadForm(10 << 20)
	if err != nil {
		return nil, nil, err
	}

	// Pegamos o primeiro arquivo dentro do campo
	fileHeaders, ok := form.File["file"]
	if !ok || len(fileHeaders) == 0 {
		return nil, nil, err
	}

	// Abrimos o arquivo
	fileHeader := fileHeaders[0]
	file, err := fileHeader.Open()

	return file, fileHeader, err
}

type SavedFileMock struct {
	mock.Mock
}

func NewSavedFileMock() *SavedFileMock {
	return &SavedFileMock{}
}

func (m *SavedFileMock) Delete() {
	m.Called()
}

func (m *SavedFileMock) Open() io.Reader {
	args := m.Called()
	return args.Get(0).(io.Reader)
}

type ReaderMock struct {
	mock.Mock
}

func NewReaderMock() *ReaderMock {
	return &ReaderMock{}
}

func (m *ReaderMock) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}
