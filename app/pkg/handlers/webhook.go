package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/ricdeau/gitlab-extension/app/pkg/broker"
	"github.com/ricdeau/gitlab-extension/app/pkg/contracts"
	"net/http"
)

// WebhookHandler handles http message from gitlab webhook pushes.
type webhookHandler struct {
	broker    broker.MessageBroker
	publishTo []string
}

// Creates new WebhookHandler instance.
func NewWebhook(broker broker.MessageBroker, publishTo ...string) HandlerFunc {
	handler := &webhookHandler{broker, publishTo}
	return func(c Context) {
		handler.handle(c)
	}
}

// Publishes http message to global broker topic.
func (handler *webhookHandler) handle(c Context) {
	logger := c.GetLogger()
	if logger == nil {
		c.SetStatusCode(http.StatusInternalServerError)
		return
	}
	var message contracts.PipelinePush
	if err := c.FromJson(&message); err != nil {
		logger.Errorf("Request body doesn't match type: %T", message)
		c.ToJson(http.StatusBadRequest, gin.H{"error": err.Error()})
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
