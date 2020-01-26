package handlers

import (
	"fmt"
	"github.com/ricdeau/gitlab-extension/app/pkg/contracts"
	"github.com/ricdeau/gitlab-extension/app/pkg/logging"
	"github.com/ricdeau/gitlab-extension/app/tests"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func TestNewWebhookHandler(t *testing.T) {
	mockBroker := new(tests.MockMessageBroker)
	actual := NewWebhook(mockBroker)
	assert.NotNil(t, actual)
	assert.IsType(t, HandlerFunc(nil), actual)
}

func TestWebhookHandler_Handle_Success(t *testing.T) {
	const (
		topic = "some topic"
		kind  = "some kind"
	)
	mockCtx := tests.DefaultMockContext()
	mockCtx.On("GetLogger").Once()
	mockCtx.On("FromJson").Once()
	mockCtx.On("SetStatusCode").Once()
	mockBroker := new(tests.MockMessageBroker)
	mockBroker.On("Publish", topic, contracts.PipelinePush{Kind: kind}).Once()
	mockLogger := new(tests.MockLogger)
	mockLogger.On("Infof").Once()
	mockCtx.BindJSON = func(m interface{}) error {
		v := reflect.ValueOf(m).Elem()
		v.Set(reflect.ValueOf(contracts.PipelinePush{Kind: kind}))
		return nil
	}
	mockCtx.Logger = func() logging.Logger {
		return mockLogger
	}

	handlerFunc := NewWebhook(mockBroker, topic)
	handlerFunc(mockCtx)

	assert.Equal(t, http.StatusOK, mockCtx.Status)
}

func TestWebhookHandler_Handle_NoTopics(t *testing.T) {
	mockCtx := tests.DefaultMockContext()
	mockCtx.On("GetLogger").Once()
	mockCtx.On("FromJson").Once()
	mockCtx.On("SetStatusCode").Once()
	mockBroker := new(tests.MockMessageBroker)
	mockLogger := new(tests.MockLogger)
	mockCtx.BindJSON = func(m interface{}) error {
		return nil
	}
	mockCtx.Logger = func() logging.Logger {
		return mockLogger
	}

	handlerFunc := NewWebhook(mockBroker)
	handlerFunc(mockCtx)

	assert.Equal(t, http.StatusOK, mockCtx.Status)
}

func TestWebhookHandler_Handle_BadRequests(t *testing.T) {
	mockCtx := tests.DefaultMockContext()
	mockCtx.On("GetLogger").Once()
	mockCtx.On("FromJson").Once()
	mockCtx.On("ToJson").Once()
	mockBroker := new(tests.MockMessageBroker)
	mockLogger := new(tests.MockLogger)
	mockLogger.On("Errorf").Once()
	mockCtx.BindJSON = func(interface{}) error {
		return fmt.Errorf("json error")
	}
	mockCtx.Logger = func() logging.Logger {
		return mockLogger
	}

	handlerFunc := NewWebhook(mockBroker)
	handlerFunc(mockCtx)

	assert.Equal(t, http.StatusBadRequest, mockCtx.Status)
}

func TestWebhookHandler_Handle_PublishError(t *testing.T) {
	const (
		topic1 = "some topic 1"
		topic2 = "some topic 2"
	)
	mockCtx := tests.DefaultMockContext()
	mockCtx.On("GetLogger").Once()
	mockCtx.On("FromJson").Once()
	mockCtx.On("SetStatusCode").Once()
	mockBroker := new(tests.MockMessageBroker)
	mockBroker.PublishError = true
	mockBroker.On("Publish", topic1, contracts.PipelinePush{}).Once()
	mockBroker.On("Publish", topic2, contracts.PipelinePush{}).Once()
	mockLogger := new(tests.MockLogger)
	mockLogger.On("Infof").Twice()
	mockLogger.On("Errorf").Twice()
	mockCtx.BindJSON = func(interface{}) error {
		return nil
	}
	mockCtx.Logger = func() logging.Logger {
		return mockLogger
	}

	handlerFunc := NewWebhook(mockBroker, topic1, topic2)
	handlerFunc(mockCtx)

	assert.Equal(t, http.StatusOK, mockCtx.Status)
}

func TestWebhookHandler_Handle_NoLogger(t *testing.T) {
	const (
		topic1 = "some topic 1"
		topic2 = "some topic 2"
	)
	mockCtx := tests.DefaultMockContext()
	mockCtx.On("GetLogger").Once()
	mockCtx.On("SetStatusCode").Once()
	mockBroker := new(tests.MockMessageBroker)

	handlerFunc := NewWebhook(mockBroker, topic1, topic2)
	handlerFunc(mockCtx)

	assert.Equal(t, http.StatusInternalServerError, mockCtx.Status)
}
