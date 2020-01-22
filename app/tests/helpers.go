package tests

import (
	"fmt"
	"github.com/ricdeau/gitlab-extension/app/pkg/broker"
	"github.com/ricdeau/gitlab-extension/app/pkg/logging"
	"github.com/stretchr/testify/mock"
)

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Infof(_ string, _ ...interface{}) {
	m.Called()
}

func (m *MockLogger) Warnf(_ string, _ ...interface{}) {
	m.Called()
}

func (m *MockLogger) Errorf(_ string, _ ...interface{}) {
	m.Called()
}

type MockMessageBroker struct {
	mock.Mock
	PublishError, SubscribeError bool
}

func (m *MockMessageBroker) AddTopic(name string) error {
	m.Called(name)
	return nil
}

func (m *MockMessageBroker) Publish(topicName string, message interface{}) error {
	m.Called(topicName, message)
	if m.PublishError {
		return fmt.Errorf("publish error")
	}
	return nil
}

func (m *MockMessageBroker) Subscribe(topicName string, consumer broker.Consumer) error {
	m.Called(topicName, consumer)
	if m.SubscribeError {
		return fmt.Errorf("subscribe error")
	}
	return nil
}

type MockContext struct {
	mock.Mock
	Status    int
	BindJSON  func(interface{}) error
	Json      func(int, interface{})
	Logger    func() logging.Logger
	SetStatus func(int)
}

func DefaultMockContext() *MockContext {
	result := &MockContext{
		Mock:      mock.Mock{},
		BindJSON:  func(interface{}) error { return nil },
		Json:      func(int, interface{}) {},
		Logger:    func() logging.Logger { return nil },
		SetStatus: nil,
	}
	result.SetStatus = func(code int) {
		result.Status = code
	}
	result.Json = func(code int, i interface{}) {
		result.Status = code
	}
	return result
}

func (m *MockContext) ShouldBindJSON(obj interface{}) error {
	m.Called()
	return m.BindJSON(obj)
}

func (m *MockContext) JSON(code int, obj interface{}) {
	m.Called()
	m.Json(code, obj)
}

func (m *MockContext) SetLogger(_ logging.Logger) {
}

func (m *MockContext) GetLogger() logging.Logger {
	m.Called()
	return m.Logger()
}

func (m *MockContext) SetStatusCode(code int) {
	m.Called()
	m.SetStatus(code)
}
