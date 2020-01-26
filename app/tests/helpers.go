package tests

import (
	"fmt"
	"github.com/ricdeau/gitlab-extension/app/pkg/broker"
	"github.com/ricdeau/gitlab-extension/app/pkg/contracts"
	"github.com/ricdeau/gitlab-extension/app/pkg/logging"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
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

func (m *MockMessageBroker) Subscribe(_ string, _ broker.Consumer) error {
	m.Called()
	if m.SubscribeError {
		return fmt.Errorf("subscribe error")
	}
	return nil
}

type MockContext struct {
	mock.Mock
	Status      int
	BindJSON    func(interface{}) error
	Json        func(int, interface{})
	Logger      func() logging.Logger
	SetStatus   func(int)
	QueryParams map[string]string
}

func (m *MockContext) QueryParam(key string) string {
	m.Called(key)
	val, exist := m.QueryParams[key]
	if exist {
		return val
	}
	return ""
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

func (m *MockContext) FromJson(obj interface{}) error {
	m.Called()
	return m.BindJSON(obj)
}

func (m *MockContext) ToJson(code int, obj interface{}) {
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

func (m *MockContext) GetWriter() http.ResponseWriter {
	m.Called()
	return new(httptest.ResponseRecorder)
}

func (m *MockContext) GetRequest() *http.Request {
	m.Called()
	return nil
}

type MockBroadcaster struct {
	mock.Mock
	BroadcastFunc     func([]byte) error
	HandleRequestFunc func(http.ResponseWriter, *http.Request) error
}

func DefaultMockBroadcaster() *MockBroadcaster {
	result := new(MockBroadcaster)
	result.BroadcastFunc = func([]byte) error {
		return nil
	}
	result.HandleRequestFunc = func(http.ResponseWriter, *http.Request) error {
		return nil
	}
	return result
}

func (m *MockBroadcaster) Broadcast(msg []byte) error {
	m.Called(msg)
	return m.BroadcastFunc(msg)
}

func (m *MockBroadcaster) HandleRequest(w http.ResponseWriter, r *http.Request) error {
	m.Called()
	return m.HandleRequestFunc(w, r)
}

type MockProjectsCache struct {
	mock.Mock
	Projects []contracts.Project
}

func (m *MockProjectsCache) GetProjects() (projects []contracts.Project, exists bool) {
	m.Called()
	if m.Projects == nil {
		return nil, false
	}
	return m.Projects, true
}

func (m *MockProjectsCache) SetProjects(projects []contracts.Project) {
	m.Called()
}

func (m *MockProjectsCache) UpdatePipeline(pipelinePush contracts.PipelinePush) error {
	m.Called(pipelinePush)
	return nil
}
