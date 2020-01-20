package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/ricdeau/gitlab-extension/app/pkg/broker"
	"github.com/ricdeau/gitlab-extension/app/pkg/contracts"
	"net/http"
)

// WebhookHandler handles http message from gitlab webhook pushes.
type WebhookHandler struct {
	broker    broker.MessageBroker
	publishTo []string
}

// Create new WebhookHandler instance.
func NewWebhookHandler(broker broker.MessageBroker, publishTo ...string) *WebhookHandler {
	return &WebhookHandler{broker, publishTo}
}

// Publishes http message to global queue topic.
func (handler *WebhookHandler) Handle(c Context) {
	logger := c.GetLogger()
	if logger == nil {
		c.SetStatusCode(http.StatusInternalServerError)
		return
	}
	var message contracts.PipelinePush
	if err := c.ShouldBindJSON(&message); err != nil {
		logger.Errorf("Request body doesn't match type: %T", message)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, topicName := range handler.publishTo {
		logger.Infof("Publishing message %+v to topic %s", message, topicName)
		if err := handler.broker.Publish(topicName, message); err != nil {
			logger.Errorf("Message publishing error: %v", err)
		}
	}
	c.SetStatusCode(http.StatusOK)
}
