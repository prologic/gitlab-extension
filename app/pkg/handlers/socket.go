package handlers

import (
	"encoding/json"
	"github.com/ricdeau/gitlab-extension/app/pkg/broker"
	"github.com/ricdeau/gitlab-extension/app/pkg/logging"
	"net/http"
)

type WsBroadcaster interface {
	Broadcast(msg []byte) error
	HandleRequest(w http.ResponseWriter, r *http.Request) error
}

// socketHandler handles messages from global broker to websockets.
type socketHandler struct {
	WsBroadcaster
	broker broker.MessageBroker
	logger logging.Logger
}

// Create new socketHandler instance
func NewSocket(topic string, broadcaster WsBroadcaster, broker broker.MessageBroker, logger logging.Logger) HandlerFunc {
	handler := &socketHandler{broadcaster, broker, logger}
	if err := handler.broker.AddTopic(topic); err != nil {
		panic(err)
	}
	err := handler.broker.Subscribe(topic, func(message interface{}) {
		msgBytes, err := json.Marshal(message)
		if err != nil {
			handler.logger.Errorf("error while marshaling message %v to json: %v", message, err)
		}
		err = handler.Broadcast(msgBytes)
		if err != nil {
			handler.logger.Errorf("websocket broadcast error on message %v: %v", message, err)
		}
	})
	if err != nil {
		panic(err)
	}
	return func(c Context) {
		handler.handle(c)
	}
}

// Handler http message. Just a stub.
func (handler *socketHandler) handle(c Context) {
	err := handler.HandleRequest(c.GetWriter(), c.GetRequest())
	if err != nil {
		handler.logger.Errorf("websocket request error: %v", err)
	}
}
