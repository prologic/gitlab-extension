package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/ricdeau/gitlab-extension/app/pkg/broker"
	"github.com/ricdeau/gitlab-extension/app/pkg/contracts"
	"github.com/sirupsen/logrus"
	"net/http"
)

// WebhookHandler handles http message from gitlab webhook pushes.
type WebhookHandler struct {
	queue     *broker.MessageBroker
	Log       *logrus.Logger
	publishTo []string
}

// Create new WebhookHandler instance.
func NewWebhookHandler(queue *broker.MessageBroker, publishTo []string, log *logrus.Logger) *WebhookHandler {
	for _, topicName := range publishTo {
		queue.AddTopic(topicName)
	}
	return &WebhookHandler{queue, log, publishTo}
}

// Publishes http message to global queue topic.
func (handler *WebhookHandler) Handle(c *gin.Context) {
	var message contracts.PipelinePush
	if err := c.ShouldBindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, topicName := range handler.publishTo {
		handler.Log.Infof("Publishing message %+v to topic %s", message, topicName)
		handler.queue.Publish(topicName, message)
	}
	c.Status(http.StatusOK)
}
