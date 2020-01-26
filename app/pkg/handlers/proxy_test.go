package handlers

import (
	"fmt"
	"github.com/ricdeau/gitlab-extension/app/pkg/config"
	"github.com/ricdeau/gitlab-extension/app/pkg/contracts"
	"github.com/ricdeau/gitlab-extension/app/pkg/logging"
	"github.com/ricdeau/gitlab-extension/app/tests"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	projId     int64 = 99
	pipelineId int64 = 999
	projName         = "project"
	branch           = "master"
	status           = "success"
	sha              = "sha999"
	title            = "commit1"
	createdAt        = "today"
	author           = "Committer"
)

func TestNewProxy(t *testing.T) {
	mockCache := new(tests.MockProjectsCache)
	mockLogger := new(tests.MockLogger)
	configMock := new(config.Config)
	actual := NewProxy(configMock, mockCache, mockLogger)
	assert.NotNil(t, actual)
	assert.IsType(t, HandlerFunc(nil), actual)
}

func TestProxyHandler_filterProjects(t *testing.T) {
	branch := "br1"
	const projId int64 = 99
	before := []contracts.Project{
		{
			Id: projId,
			Pipelines: []contracts.Pipeline{
				{
					Id:     999,
					Branch: branch,
				},
				{
					Id:     88,
					Branch: "br2",
				},
			},
		},
		{
			Id: 10,
		},
	}

	assert.Equal(t, 2, len(before))
	assert.Equal(t, 2, len(before[0].Pipelines))

	ids := make(map[int64]struct{})
	ids[projId] = struct{}{}
	branches := make(map[string]struct{})
	branches[branch] = struct{}{}
	after := filterProjects(before, ids, branches)

	assert.Equal(t, 1, len(after))
	assert.Equal(t, projId, after[0].Id)
	assert.Equal(t, 1, len(after[0].Pipelines))
	assert.Equal(t, branch, after[0].Pipelines[0].Branch)
}

func TestProxyHandler_getCommitForProject(t *testing.T) {
	ts := createTestServer()
	defer ts.Close()
	client := ts.Client()
	mockLogger := new(tests.MockLogger)
	mockLogger.On("Infof").Twice()
	mockLogger.On("Errorf").Once()
	configMock := new(config.Config)
	handler := &proxyHandler{config: configMock, client: client, gitlabUrl: ts.URL}

	actual, err := handler.getCommitForProject(projId, sha, mockLogger)
	if assert.NoError(t, err) {
		assert.NotNil(t, actual)
		assert.Equal(t, &contracts.Commit{
			Title:     title,
			CreatedAt: createdAt,
			Author:    author,
		}, actual)
	}
	_, err = handler.getCommitForProject(0, sha, mockLogger)
	assert.Error(t, err)
}

func TestProxyHandler_getPipelines(t *testing.T) {
	ts := createTestServer()
	defer ts.Close()
	client := ts.Client()
	mockLogger := new(tests.MockLogger)
	mockLogger.On("Infof")
	mockLogger.On("Errorf").Once()
	configMock := new(config.Config)
	handler := &proxyHandler{config: configMock, client: client, gitlabUrl: ts.URL}

	actual, err := handler.getPipelines(projId, 1, mockLogger)
	if assert.NoError(t, err) {
		assert.NotNil(t, actual)
		assert.Equal(t, 1, len(actual))
		assert.Equal(t, pipelineId, actual[0].Id)
		assert.Equal(t, sha, actual[0].Sha)
		assert.Equal(t, branch, actual[0].Branch)
		assert.Equal(t, status, actual[0].Status)
	}
	_, err = handler.getPipelines(0, 1, mockLogger)
	assert.Error(t, err)
}

func TestProxyHandler_getProjects(t *testing.T) {
	ts := createTestServer()
	defer ts.Close()
	client := ts.Client()
	mockLogger := new(tests.MockLogger)
	mockLogger.On("Infof")
	configMock := new(config.Config)
	mockCache := new(tests.MockProjectsCache)
	mockCache.On("GetProjects").Once()
	mockCache.On("SetProjects").Once()
	handler := &proxyHandler{
		config:    configMock,
		client:    client,
		gitlabUrl: ts.URL,
		cache:     mockCache,
	}

	actual, err := handler.getProjects(1, mockLogger)
	if assert.NoError(t, err) {
		assert.NotNil(t, actual)
		assert.Equal(t, 1, len(actual))
		assert.Equal(t, projId, actual[0].Id)
		assert.Equal(t, projName, actual[0].Name)
		assert.Equal(t, 1, len(actual[0].Pipelines))
		assert.Equal(t, pipelineId, actual[0].Pipelines[0].Id)
	}
}

func TestProxyHandler_handle(t *testing.T) {
	const (
		idsParam      = "project_ids"
		branchesParam = "branches"
	)
	ts := createTestServer()
	defer ts.Close()
	client := ts.Client()
	mockLogger := new(tests.MockLogger)
	mockLogger.On("Infof")

	configMock := new(config.Config)

	mockCache := new(tests.MockProjectsCache)
	mockCache.On("GetProjects").Once()
	mockCache.On("SetProjects").Once()

	mockContext := tests.DefaultMockContext()
	mockContext.QueryParams = make(map[string]string)
	projStr := fmt.Sprintf("%d", projId)
	mockContext.QueryParams[idsParam] = projStr
	mockContext.QueryParams[branchesParam] = branch
	mockContext.Logger = func() logging.Logger {
		return mockLogger
	}
	mockContext.On("GetLogger").Once()
	mockContext.On("QueryParam", idsParam).Once()
	mockContext.On("QueryParam", branchesParam).Once()
	mockContext.On("ToJson").Once()
	handler := &proxyHandler{
		config:    configMock,
		client:    client,
		gitlabUrl: ts.URL,
		cache:     mockCache,
	}
	handler.handle(mockContext)
}

func createTestServer() *httptest.Server {
	const commitResponseFormat = `{
									"title" : "%s",
									"created_at" : "%s",
									"author_name" : "%s"
								  }`
	const pipelineResponseFormat = `[{
									  "id" : %d,
									  "sha" : "%s",
									  "ref" : "%s",
									  "status" : "%s",
									  "web_url" : "url"
									}]`
	const projectsResponseFormat = `[{
										"id" : %d,
										"name" : "%s",
										"namespace" : {
											"name" : "ns"
										},
										"last_activity_at" : "today",
										"web_url" : "url"
									}]`

	r := http.NewServeMux()
	r.HandleFunc(fmt.Sprintf(commitUrl, "", projId, sha), func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, commitResponseFormat, title, createdAt, author)
		if err != nil {
			w.WriteHeader(500)
		}
	})
	r.HandleFunc(fmt.Sprintf(pipelinesUtl, "", projId), func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, pipelineResponseFormat, pipelineId, sha, branch, status)
		if err != nil {
			w.WriteHeader(500)
		}
	})
	r.HandleFunc(fmt.Sprintf(projectsUrl, ""), func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintf(w, projectsResponseFormat, projId, projName)
		if err != nil {
			w.WriteHeader(500)
		}
	})
	return httptest.NewServer(r)
}
