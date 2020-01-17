package queue

import (
	"github.com/sirupsen/logrus"
	"sync"
)

type Consumer func(interface{})

// GlobalQueue consists of several topics,
// consumers can subscribe on them.
type GlobalQueue struct {
	topics map[string]chan interface{}
	lock   *sync.Mutex
	log    *logrus.Logger
}

// Returns reference on new GlobalQueue instance.
func NewGlobalQueue(logger *logrus.Logger) *GlobalQueue {
	result := GlobalQueue{}
	result.log = logger
	result.topics = make(map[string]chan interface{})
	result.lock = new(sync.Mutex)
	return &result
}

// Adds new topic in queue, if topic with given name exists nothing will happen.
func (q *GlobalQueue) AddTopic(name string) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if _, ok := q.topics[name]; !ok {
		q.topics[name] = make(chan interface{})
	}
}

// Publishes message in topic.
func (q *GlobalQueue) Publish(topicName string, message interface{}) {
	topic, ok := q.topics[topicName]
	if ok {
		topic <- message
	} else {
		q.log.Errorf("error publish: there is no topic named '%s'", topicName)
	}
}

// Binds consuming functions to queue topic.
func (q *GlobalQueue) Subscribe(topicName string, consumer Consumer) {
	topic, ok := q.topics[topicName]
	if ok {
		go func() {
			for message := range topic {
				consumer(message)
			}
		}()
	} else {
		q.log.Errorf("error subscribe: there is no topic named '%s'", topicName)
	}
}
