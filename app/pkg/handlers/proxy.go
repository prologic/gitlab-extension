package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/ricdeau/gitlab-extension/app/pkg/caching"
	"github.com/ricdeau/gitlab-extension/app/pkg/config"
	"github.com/ricdeau/gitlab-extension/app/pkg/contracts"
	"github.com/ricdeau/gitlab-extension/app/pkg/logging"
	"github.com/ricdeau/gitlab-extension/app/pkg/utils"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	privateToken    = "Private-Token"
	pipelinesNumber = 5
)

// urls
const (
	projectsUrl  = "%s/projects"
	pipelinesUtl = "%s/projects/%d/pipelines"
	commitUrl    = "%s/projects/%d/repository/commits/%s"
)

// proxyHandler that performs multiple requests to gitlab API and returns single combined response.
// with all projects, first N pipelines for each project, and last commit for each pipeline.
type proxyHandler struct {
	config    *config.Config
	logger    logging.Logger
	gitlabUrl string
	client    *http.Client
	cache     caching.ProjectsCache
}

// Create new instance of proxyHandler.
// config - Global config
// cache - Caching module
// logger - Logging module
func NewProxy(conf *config.Config, cache caching.ProjectsCache, logger logging.Logger) HandlerFunc {
	handler := &proxyHandler{}
	handler.config = conf
	handler.cache = cache
	handler.logger = logger
	handler.gitlabUrl = conf.GitlabUri
	handler.client = &http.Client{
		Timeout: time.Second * 30,
	}
	return func(c Context) {
		handler.handle(c)
	}
}

// CreateHandler '/projects' request
func (handler *proxyHandler) handle(c Context) {
	logger := c.GetLogger()
	if logger == nil {
		logger = handler.logger
	}

	// parse project ids
	projectIdsParam := c.QueryParam("project_ids")
	projectIds := make(map[int64]struct{})
	if projectIdsParam != "" {
		for _, idParam := range strings.Split(projectIdsParam, " ") {
			id, err := strconv.ParseInt(idParam, 10, 64)
			if err != nil {
				logger.Warnf("Invalid projectId %s will be skipped", idParam)
				continue
			}
			projectIds[id] = struct{}{}
		}
	}

	// parse branches
	branchesParam := c.QueryParam("branches")
	branches := make(map[string]struct{})
	if branchesParam != "" {
		for _, branchParam := range strings.Split(branchesParam, " ") {
			branches[branchParam] = struct{}{}
		}
	}

	projects, err := handler.getProjects(pipelinesNumber, logger)
	if err != nil {
		c.ToJson(http.StatusInternalServerError, contracts.NewErrorResponse(err))
		return
	}

	// filter projects by ids and pipelines by branches
	projects = filterProjects(projects, projectIds, branches)
	c.ToJson(200, contracts.NewProjectsResponse(projects))
}

// Filter projects by provided ids and filter each project pipelines by provided branches
// projects - ProjectsResponse structure from gitlab API
// projectIds - Ids of gitlab projects to filter
// branches - Branches to filter pipelines
func filterProjects(projects []contracts.Project, projectIds map[int64]struct{}, branches map[string]struct{}) (result []contracts.Project) {
	for _, project := range projects {
		if len(projectIds) != 0 {
			_, exist := projectIds[project.Id]
			if !exist {
				continue
			}
		}
		var filteredPipelines []contracts.Pipeline

		for _, pipe := range project.Pipelines {
			if len(branches) != 0 {
				_, exist := branches[pipe.Branch]
				if !exist {
					continue
				}
				filteredPipelines = append(filteredPipelines, pipe)
			}
		}
		project.Pipelines = filteredPipelines
		result = append(result, project)
	}
	return
}

// Gets all projects, allowed for private token that provided through proxyHandler.Config.
// ProjectsResponse will be cached, if cache is empty or expired, http request will be processed.
// nPipelines - top N pipelines to take
func (handler *proxyHandler) getProjects(
	nPipelines int,
	logger logging.Logger) (result []contracts.Project, err error) {

	// return if cached
	exists := false
	result, exists = handler.cache.GetProjects()
	if exists {
		return
	}

	// get new if not found in cache
	url := fmt.Sprintf("%s/projects", handler.gitlabUrl)
	response, err := handler.performGetRequest(url, logger)
	if err != nil {
		return
	}
	defer response.Body.Close()
	var rawJson []map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&rawJson)
	if err != nil {
		return
	}

	results := make(chan contracts.Project)
	go func() {
		sema := utils.CountingSemaphore{Count: 4}
		for _, projRaw := range rawJson {
			sema.Acquire()
			go func(p map[string]interface{}) {
				defer sema.Release()
				project := contracts.Project{
					Id:           int64(p["id"].(float64)),
					Name:         p["name"].(string),
					Namespace:    p["namespace"].(map[string]interface{})["name"].(string),
					LastActivity: p["last_activity_at"].(string),
					WebUrl:       p["web_url"].(string),
				}
				// add pipelines to project
				project.Pipelines, err = handler.getPipelines(project.Id, nPipelines, logger)
				if err != nil {
					logger.Errorf("ErrorResponse while getting pipelines: %v", err)
				}
				results <- project
			}(projRaw)
		}
		sema.WaitAll()
		close(results)
	}()

	for r := range results {
		result = append(result, r)
	}

	handler.cache.SetProjects(result)
	return
}

// Gets pipelines for project.
// projectId - the identifier of gitlab project
// nPipelines - top N pipelines to take
func (handler *proxyHandler) getPipelines(
	projectId int64,
	nPipelines int,
	logger logging.Logger) (pipelines []contracts.Pipeline, err error) {
	url := fmt.Sprintf(pipelinesUtl, handler.gitlabUrl, projectId)
	response, err := handler.performGetRequest(url, logger)
	if err != nil {
		return
	}
	defer response.Body.Close()
	var rawJson []map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&rawJson)
	if err != nil {
		return
	}
	if nPipelines > len(rawJson) {
		nPipelines = len(rawJson)
	}
	for _, p := range rawJson[:nPipelines] {
		pipeline := contracts.Pipeline{
			Id:     int64(p["id"].(float64)),
			Sha:    p["sha"].(string),
			Branch: p["ref"].(string),
			Status: p["status"].(string),
			WebUrl: p["web_url"].(string),
		}
		// add last commit to pipeline
		pipeline.Commit, err = handler.getCommitForProject(projectId, pipeline.Sha, logger)
		if err != nil {
			return
		}
		pipelines = append(pipelines, pipeline)
	}
	return
}

// Gets commit and converts it to contracts.Commit struct.
// projectId - the identifier of gitlab project
// sha - commit's SHA
func (handler *proxyHandler) getCommitForProject(
	projectId int64,
	sha string,
	logger logging.Logger) (result *contracts.Commit, err error) {
	url := fmt.Sprintf(commitUrl, handler.gitlabUrl, projectId, sha)
	response, err := handler.performGetRequest(url, logger)
	if err != nil {
		return
	}
	defer response.Body.Close()
	var rawJson map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&rawJson)
	if err != nil {
		return
	}
	result = &contracts.Commit{
		Title:     rawJson["title"].(string),
		CreatedAt: rawJson["created_at"].(string),
		Author:    rawJson["author_name"].(string),
	}
	return
}

// Performs GET request with Private-Token header and returns response.
// url - request's url
func (handler *proxyHandler) performGetRequest(url string, logger logging.Logger) (*http.Response, error) {
	headers := map[string]string{privateToken: handler.config.GitlabToken}
	return utils.PerformGetRequest(handler.client, url, headers, logger)
}
