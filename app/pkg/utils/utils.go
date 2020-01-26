package utils

import (
	"fmt"
	"github.com/ricdeau/gitlab-extension/app/pkg/logging"
	"net/http"
	"sync"
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

type CountingSemaphore struct {
	Count   int
	wg      sync.WaitGroup
	once    sync.Once
	counter chan struct{}
}

func (s *CountingSemaphore) Acquire() {
	if s.Count == 0 {
		panic("CountingSemaphore counter is 0")
	}
	if s.counter == nil {
		s.once.Do(func() {
			s.counter = make(chan struct{}, s.Count)
		})
	}
	s.wg.Add(1)
	s.counter <- struct{}{}
}

func (s *CountingSemaphore) Release() {
	<-s.counter
	s.wg.Done()
}

func (s *CountingSemaphore) WaitAll() {
	s.wg.Wait()
}
