package broker

import (
	"fmt"
	"sync"
)

const (
	consumerIsNil    = "consumer can't be nil"
	topicNameIsEmpty = "topic name can't be empty"
	noTopic          = "there is no topic named '%s'"
	publishNoTopic   = "publish: " + noTopic
	subscribeNoTopic = "subscribe: " + noTopic
)

type Consumer func(interface{})

// MessageBroker consists of several topics,
// consumers can subscribe on them.
type MessageBroker struct {
	topics map[string]chan interface{}
	lock   *sync.Mutex
}

// Returns pointer to new MessageBroker instance.
func New() *MessageBroker {
	result := MessageBroker{}
	result.topics = make(map[string]chan interface{})
	result.lock = new(sync.Mutex)
	return &result
}

// Adds new topic in queue, if topic with given name exists nothing will happen.
func (b *MessageBroker) AddTopic(name string) error {
	if name == "" {
		return fmt.Errorf(topicNameIsEmpty)
	}
	b.lock.Lock()
	defer b.lock.Unlock()
	if _, ok := b.topics[name]; !ok {
		b.topics[name] = make(chan interface{})
	}
	return nil
}

// Publishes message in topic.
// Blocks if subscriber for this topicName hasn't been set.
func (b *MessageBroker) Publish(topicName string, message interface{}) error {
	topic, ok := b.topics[topicName]
	if ok {
		topic <- message
	} else {
		return fmt.Errorf(publishNoTopic, topicName)
	}
	return nil
}

// Binds consuming functions to queue topic.
func (b *MessageBroker) Subscribe(topicName string, consumer Consumer) error {
	if consumer == nil {
		return fmt.Errorf(consumerIsNil)
	}
	topic, ok := b.topics[topicName]
	if ok {
		go func() {
			for message := range topic {
				consumer(message)
			}
		}()
	} else {
		return fmt.Errorf(subscribeNoTopic, topicName)
	}
	return nil
}
