package handlers

import (
	"github.com/ricdeau/gitlab-extension/app/tests"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSocket(t *testing.T) {
	mockBroker := new(tests.MockMessageBroker)
	mockBroker.On("AddTopic", "topic")
	mockBroker.On("Subscribe").Once()
	mockBroadcaster := tests.DefaultMockBroadcaster()
	mockLogger := new(tests.MockLogger)
	actual := NewSocket("topic", mockBroadcaster, mockBroker, mockLogger)
	assert.NotNil(t, actual)
	assert.IsType(t, HandlerFunc(nil), actual)
}
