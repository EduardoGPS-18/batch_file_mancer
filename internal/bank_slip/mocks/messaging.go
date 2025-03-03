package bank_slip

import (
	"github.com/stretchr/testify/mock"
)

type KafkaMessageMock struct {
	mock.Mock
}

func NewKafkaMessageMock() *KafkaMessageMock {
	return &KafkaMessageMock{}
}

func (m *KafkaMessageMock) Topic() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *KafkaMessageMock) Data() (map[string]any, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *KafkaMessageMock) Commit() {
	m.Called()
}
