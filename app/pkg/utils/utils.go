package utils

import (
	"fmt"
	"github.com/ricdeau/gitlab-extension/app/pkg/logging"
	"net/http"
)

// PerformGetRequest - performs GET request with Private-Token header and returns response.
// client - http client to perform request
// url - request's url
// headers - request's headers map
func PerformGetRequest(
	client *http.Client,
	url string,
	headers map[string]string,
	logger logging.Logger) (resp *http.Response, err error) {

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	reqString := fmt.Sprintf("(Method: %s, Path: %s, Headers: %s)", request.Method, request.URL, request.Header)
	logger.Infof("Request: %s", reqString)
	resp, err = client.Do(request)
	switch {
	case err != nil:
		logger.Errorf("Error for request %s: %v", reqString, err)
	case resp.StatusCode > 299:
		logger.Errorf("Unexpected status code: %d", resp.StatusCode)
		err = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return
}
