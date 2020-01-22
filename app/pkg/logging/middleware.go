package logging

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"io/ioutil"
	"math"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	CorrelationIdKey = "correlationId"
)

// Intermediate response logger
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Middleware is a logrus logger handler
func Middleware(logger logrus.FieldLogger) gin.HandlerFunc {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return func(c *gin.Context) {
		correlationId := uuid.New()
		path := c.Request.URL.Path
		start := time.Now()
		var body interface{}
		b, _ := ioutil.ReadAll(c.Request.Body)
		rdr1 := ioutil.NopCloser(bytes.NewBuffer(b))
		rdr2 := ioutil.NopCloser(bytes.NewBuffer(b))
		_ = json.NewDecoder(rdr1).Decode(&body)
		c.Request.Body = rdr2

		entry := logger.WithFields(logrus.Fields{
			CorrelationIdKey: correlationId,
			"hostname":       hostname,
			"method":         c.Request.Method,
			"path":           path,
			"body":           body,
		})

		c.Set(CorrelationIdKey, correlationId)

		entry.Info("[Request]")

		body = nil
		buf := new(bytes.Buffer)
		writer := &responseWriter{body: buf, ResponseWriter: c.Writer}
		c.Writer = writer

		c.Next()

		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		statusCode := c.Writer.Status()
		dataLength := c.Writer.Size()
		_ = json.NewDecoder(writer.body).Decode(&body)

		entry = logger.WithFields(logrus.Fields{
			"correlationId": correlationId,
			"hostname":      hostname,
			"method":        c.Request.Method,
			"path":          path,
			"statusCode":    statusCode,
			"latency":       latency, // time to process
			"dataLength":    dataLength,
			"body":          body,
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			msg := "[Response]"
			if statusCode > 499 {
				entry.Error(msg)
			} else if statusCode > 399 {
				entry.Warn(msg)
			} else {
				entry.Info(msg)
			}
		}
	}
}
