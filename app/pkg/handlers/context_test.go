package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/ricdeau/gitlab-extension/app/tests"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHandlerFunc_CreateHandler(t *testing.T) {
	var handlerFunc HandlerFunc
	handlerFunc = func(Context) {}
	ginHandler := handlerFunc.CreateHandler()

	assert.NotNil(t, ginHandler)
	assert.IsType(t, gin.HandlerFunc(nil), ginHandler)
}

func TestGinContext_GetSetLogger(t *testing.T) {
	var context Context
	context = &GinContext{new(gin.Context)}
	mockLogger := new(tests.MockLogger)

	before := context.GetLogger()
	assert.Nil(t, before)

	context.SetLogger(mockLogger)

	after := context.GetLogger()
	assert.NotNil(t, after)
}
