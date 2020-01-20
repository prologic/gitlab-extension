package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/ricdeau/gitlab-extension/app/pkg/utils"
)

const eventLogger = "eventLogger"

type Context interface {
	ShouldBindJSON(obj interface{}) error
	JSON(code int, obj interface{})
	SetLogger(logger utils.Logger)
	GetLogger() utils.Logger
	SetStatusCode(code int)
}

type GinContext struct {
	gin.Context
}

func (c *GinContext) SetLogger(logger utils.Logger) {
	c.Set(eventLogger, logger)
}

func (c *GinContext) GetLogger() utils.Logger {
	logger, exists := c.Get(eventLogger)
	if exists {
		result := logger.(utils.Logger)
		return result
	}
	return nil
}

func (c *GinContext) SetStatusCode(code int) {
	c.Status(code)
}
