package broker

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const topic1 = "topic1"

func TestMessageBroker_AddTopic_Success(t *testing.T) {
	b := New()
	err := b.AddTopic(topic1)
	if assert.NoError(t, err) {
		topics := b.(*messageBroker).topics
		assert.Equal(t, 1, len(topics))
		assert.Contains(t, topics, topic1)
	}
}

func TestMessageBroker_AddTopic_Error(t *testing.T) {
	var topicName string
	b := New()
	err := b.AddTopic(topicName)
	if assert.EqualError(t, err, topicNameIsEmpty) {
		topics := b.(*messageBroker).topics
		assert.Empty(t, topics)
	}
}

func TestMessageBroker_PubSub_Success(t *testing.T) {
	actual := "test"
	var expected interface{}
	b := New()
	err := b.AddTopic(topic1)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	err = b.Subscribe(topic1, func(msg interface{}) {
		expected = msg
		cancel()
	})
	assert.NoError(t, err)

	err = b.Publish(topic1, actual)
	assert.NoError(t, err)

	<-ctx.Done()
	assert.Equal(t, expected, actual)
}

func TestMessageBroker_Publish_NoTopic(t *testing.T) {
	b := New()
	actualErr := b.Publish(topic1, "test")
	expectedErr := fmt.Sprintf(publishNoTopic, topic1)
	assert.EqualError(t, actualErr, expectedErr)
}

func TestMessageBroker_Subscribe_NilConsumer(t *testing.T) {
	b := New()
	actualErr := b.Subscribe(topic1, nil)
	assert.EqualError(t, actualErr, consumerIsNil)
}

func TestMessageBroker_Subscribe_NoTopic(t *testing.T) {
	b := New()
	consumer := func(interface{}) {}
	actualErr := b.Subscribe(topic1, consumer)
	expectedErr := fmt.Sprintf(subscribeNoTopic, topic1)
	assert.EqualError(t, actualErr, expectedErr)
}
