package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/ricdeau/gitlab-extension/app/pkg/logging"
	"net/http"
)

const eventLogger = "eventLogger"

type HandlerFunc func(Context)

func (h HandlerFunc) CreateHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		h(&GinContext{ctx})
	}
}

type Context interface {
	ShouldBindJSON(obj interface{}) error
	JSON(code int, obj interface{})
	SetLogger(logger logging.Logger)
	GetLogger() logging.Logger
	SetStatusCode(code int)
	GetWriter() http.ResponseWriter
	GetRequest() *http.Request
}

type GinContext struct {
	*gin.Context
}

func (c *GinContext) SetLogger(logger logging.Logger) {
	c.Set(eventLogger, logger)
}

func (c *GinContext) GetLogger() logging.Logger {
	logger, exists := c.Get(eventLogger)
	if exists {
		result := logger.(logging.Logger)
		return result
	}
	return nil
}

func (c *GinContext) SetStatusCode(code int) {
	c.Status(code)
}

func (c *GinContext) GetWriter() http.ResponseWriter {
	return c.Writer
}

func (c *GinContext) GetRequest() *http.Request {
	return c.Request
}
