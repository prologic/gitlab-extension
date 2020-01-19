package handlers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/ricdeau/gitlab-extension/app/pkg/broker"
	"github.com/sirupsen/logrus"
	"gopkg.in/olahol/melody.v1"
)

const (
	SocketTopic = "ws"
)

// SocketHandler handles messages from global queue to websockets.
type SocketHandler struct {
	*melody.Melody
	queue broker.MessageBroker
	Log   *logrus.Logger
}

// Create new SocketHandler instance
func NewSocketHandler(wsHandler *melody.Melody, queue broker.MessageBroker, logger *logrus.Logger) *SocketHandler {
	handler := SocketHandler{wsHandler, queue, logger}
	handler.queue.AddTopic(SocketTopic)
	handler.queue.Subscribe(SocketTopic, func(message interface{}) {
		msgBytes, err := json.Marshal(message)
		if err != nil {
			handler.Log.Errorf("error while marshaling message %v to json: %v", message, err)
		}
		err = handler.Broadcast(msgBytes)
		if err != nil {
			handler.Log.Errorf("websocket broadcast error on message %v: %v", message, err)
		}
	})
	return &handler
}

// Handle http message. Just a stub.
func (handler *SocketHandler) Handle(c *gin.Context) {
	err := handler.HandleRequest(c.Writer, c.Request)
	if err != nil {
		handler.Log.Errorf("websocket request error: %v", err)
	}
}
