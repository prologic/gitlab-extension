package utils

import (
	"github.com/ricdeau/gitlab-extension/app/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestPerformGetRequest_Success(t *testing.T) {
	expected := []byte("success")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(expected)
	}))
	defer ts.Close()

	logger := new(tests.MockLogger)
	logger.On("Infof").Once()

	client := ts.Client()
	headers := make(map[string]string)
	headers["Timeout"] = "10"
	resp, err := PerformGetRequest(client, ts.URL, headers, logger)

	mock.AssertExpectationsForObjects(t, logger)

	if assert.NoError(t, err) {
		assert.NotNil(t, resp, "response is nil")
		assert.Equalf(t, 200, resp.StatusCode, "request status code != 200, actual code is %d", resp.StatusCode)
		actual, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if assert.NoError(t, err) {
			assert.Equal(t, expected, actual, "response body does't match")
		}
	}
}

func TestPerformGetRequest_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	logger := new(tests.MockLogger)
	logger.On("Infof").Once()
	logger.On("Errorf").Once()

	client := ts.Client()
	ts.Close()
	headers := make(map[string]string)
	headers["Timeout"] = "10"
	_, err := PerformGetRequest(client, ts.URL, headers, logger)

	mock.AssertExpectationsForObjects(t, logger)
	if assert.Error(t, err) {
		urlErr, ok := err.(*url.Error)
		if !ok {
			assert.Fail(t, "error is not of type url.Error")
		}
		var expected *net.OpError
		assert.IsType(t, expected, urlErr.Err)
	}
}

func TestPerformGetRequest_BadStatusCode(t *testing.T) {
	statusCode := 404
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}))
	defer ts.Close()

	logger := new(tests.MockLogger)
	logger.On("Infof").Once()
	logger.On("Errorf").Once()

	client := ts.Client()
	headers := make(map[string]string)
	headers["Timeout"] = "10"
	_, err := PerformGetRequest(client, ts.URL, headers, logger)

	mock.AssertExpectationsForObjects(t, logger)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), strconv.Itoa(statusCode))
	}
}
