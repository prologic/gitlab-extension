package utils

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
)

// Performs GET request with Private-Token header and returns response.
// client - http client to perform request
// url - request's url
// headers - request's headers collection
func PerformGetRequest(
	client *http.Client,
	url string,
	headers map[string]string,
	logger *logrus.Entry) (*http.Response, error) {

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	reqString := fmt.Sprintf("(Method: %s, Path: %s, Headers: %s)", request.Method, request.URL, request.Header)
	logger.Infof("Request: %s", reqString)
	resp, err := client.Do(request)
	if err != nil {
		logger.Errorf("Error for request %s: %v", reqString, err)
		return nil, err
	}
	if resp.StatusCode > 299 {
		logger.Warnf("Unexpected status code: %d", resp.StatusCode)
	}
	return resp, nil
}
